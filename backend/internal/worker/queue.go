package worker

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/config"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/repository"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/storage"
)

type Queue struct {
	jobs      *repository.JobRepo
	versions  *repository.VersionRepo
	entries   *repository.EntryRepo
	sets      *repository.SetRepo
	storage   *storage.COSStorage
	inspector BundleInspector
	tempDir   string
	maxZipMB  int
	jobCh     chan int64
}

func NewQueue(
	jobs *repository.JobRepo,
	versions *repository.VersionRepo,
	entries *repository.EntryRepo,
	sets *repository.SetRepo,
	storage *storage.COSStorage,
	inspector BundleInspector,
	cfg *config.Config,
) *Queue {
	q := &Queue{
		jobs:      jobs,
		versions:  versions,
		entries:   entries,
		sets:      sets,
		storage:   storage,
		inspector: inspector,
		tempDir:   cfg.Upload.TempDir,
		maxZipMB:  cfg.Upload.MaxZipMB,
		jobCh:     make(chan int64, 100),
	}
	go q.loop()
	return q
}

func (q *Queue) Enqueue(jobID int64) {
	q.jobCh <- jobID
}

func (q *Queue) loop() {
	for jobID := range q.jobCh {
		if err := q.process(context.Background(), jobID); err != nil {
			log.Printf("job %d failed: %v", jobID, err)
		}
	}
}

func (q *Queue) process(ctx context.Context, jobID int64) error {
	job, err := q.jobs.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if err := q.jobs.UpdateStatus(ctx, jobID, domain.JobStatusRunning, ""); err != nil {
		return err
	}

	version, err := q.versions.GetByID(ctx, job.VersionID)
	if err != nil {
		return q.fail(ctx, jobID, job.VersionID, err)
	}
	set, err := q.sets.GetByID(ctx, version.SetID)
	if err != nil {
		return q.fail(ctx, jobID, job.VersionID, err)
	}

	workDir := job.TempDir
	if workDir == "" {
		workDir = filepath.Join(q.tempDir, fmt.Sprintf("job_%d", jobID))
	}

	defer func() {
		_ = os.RemoveAll(workDir)
	}()

	q.updateProgress(ctx, version.ID, domain.ParseStatusRunning, "unzipping", 15, "Mengekstrak paket zip…")

	zipPath := filepath.Join(workDir, "upload.zip")
	if err := unzipSafe(zipPath, workDir); err != nil {
		return q.fail(ctx, jobID, version.ID, fmt.Errorf("unzip: %w", err))
	}

	q.updateProgress(ctx, version.ID, domain.ParseStatusRunning, "validating", 30, "Memvalidasi struktur paket…")
	meta, err := ValidateExtracted(workDir, set.Slug)
	if err != nil {
		return q.fail(ctx, jobID, version.ID, err)
	}
	if version.Version != "" && meta.Version != "" && version.Version != meta.Version {
		return q.fail(ctx, jobID, version.ID, fmt.Errorf("version mismatch: zip=%s bundle=%s", version.Version, meta.Version))
	}
	if meta.Version != "" {
		version.Version = meta.Version
	}

	q.updateProgress(ctx, version.ID, domain.ParseStatusRunning, "inspecting", 50, "Memeriksa info AssetBundle…")
	inspectDir := filepath.Join(workDir, "inspect")
	manifest, err := q.inspector.Inspect(ctx, meta.InfoPath, inspectDir, set.Slug)
	if err != nil {
		return q.fail(ctx, jobID, version.ID, err)
	}
	if manifest.Version != "" && version.Version != "" && manifest.Version != version.Version {
		return q.fail(ctx, jobID, version.ID, fmt.Errorf("version mismatch: filename=%s bundle=%s", version.Version, manifest.Version))
	}
	if manifest.Version != "" {
		version.Version = manifest.Version
	}

	q.updateProgress(ctx, version.ID, domain.ParseStatusRunning, "uploading", 70, "Mengunggah ke COS…")

	zipKey := storage.PackageKey(set.AuthorID, set.Slug, version.Version, fmt.Sprintf("%s_v%s_%s.zip", set.Slug, version.Version, time.Now().Format("20060102")))
	if err := q.storage.UploadFile(ctx, zipKey, zipPath); err != nil {
		return q.fail(ctx, jobID, version.ID, fmt.Errorf("upload zip: %w", err))
	}
	version.ZipCOSKey = zipKey
	version.FileCount = meta.FileCount
	version.HasRuntime = meta.HasRuntime
	version.HasCommonW2 = meta.HasCommonW2
	version.UID = manifest.UID

	var levelEntries []domain.LevelEntry
	for i, lv := range manifest.Levels {
		entry := domain.LevelEntry{
			VersionID:   version.ID,
			LevelID:     lv.ID,
			LevelName:   lv.LevelName,
			LevelNameZH: lv.LevelNameZH,
			SceneName:   lv.SceneName,
			SortOrder:   i,
		}
		if lv.Screenshot != "" {
			localShot := filepath.Join(inspectDir, lv.Screenshot)
			if _, err := os.Stat(localShot); err == nil {
				shotKey := storage.ScreenshotKey(set.AuthorID, set.Slug, version.Version, lv.ID)
				if err := q.storage.UploadFile(ctx, shotKey, localShot); err == nil {
					entry.ScreenshotCOSKey = shotKey
				}
			}
		}
		levelEntries = append(levelEntries, entry)
	}

	if len(levelEntries) > 0 && levelEntries[0].ScreenshotCOSKey != "" {
		coverKey := storage.CoverKey(set.Slug)
		localCover := filepath.Join(inspectDir, manifest.Levels[0].Screenshot)
		if err := q.storage.UploadFile(ctx, coverKey, localCover); err == nil {
			set.CoverCOSKey = coverKey
			set.Name = manifest.LevelSetName
			set.NameZH = manifest.LevelSetNameZH
			_ = q.sets.Update(ctx, set)
		}
	}

	q.updateProgress(ctx, version.ID, domain.ParseStatusRunning, "finalizing", 90, "Menulis ke database…")

	if err := q.entries.CreateBatch(ctx, levelEntries); err != nil {
		return q.fail(ctx, jobID, version.ID, err)
	}

	version.ParseMessage = "Parsing selesai"
	if err := q.versions.Finalize(ctx, version); err != nil {
		return q.fail(ctx, jobID, version.ID, err)
	}

	_ = q.jobs.UpdateStatus(ctx, jobID, domain.JobStatusDone, "")
	return nil
}

func (q *Queue) updateProgress(ctx context.Context, versionID int64, status domain.ParseStatus, phase string, progress int, msg string) {
	_ = q.versions.UpdateProgress(ctx, versionID, status, phase, progress, msg)
}

func (q *Queue) fail(ctx context.Context, jobID, versionID int64, err error) error {
	msg := err.Error()
	_ = q.versions.MarkFailed(ctx, versionID, msg)
	_ = q.jobs.UpdateStatus(ctx, jobID, domain.JobStatusFailed, msg)
	return err
}

func unzipSafe(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	dest = filepath.Clean(dest)
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(filepath.Clean(target), dest+string(os.PathSeparator)) && filepath.Clean(target) != dest {
			return fmt.Errorf("zip path traversal: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		out.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
