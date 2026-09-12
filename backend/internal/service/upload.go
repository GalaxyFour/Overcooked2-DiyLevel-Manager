package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/repository"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/worker"
)

type UploadService struct {
	sets     *SetService
	versions *repository.VersionRepo
	jobs     *repository.JobRepo
	queue    *worker.Queue
	tempDir  string
	maxZipMB int
}

func NewUploadService(
	sets *SetService,
	versions *repository.VersionRepo,
	jobs *repository.JobRepo,
	queue *worker.Queue,
	tempDir string,
	maxZipMB int,
) *UploadService {
	return &UploadService{
		sets:     sets,
		versions: versions,
		jobs:     jobs,
		queue:    queue,
		tempDir:  tempDir,
		maxZipMB: maxZipMB,
	}
}

func (s *UploadService) Upload(ctx context.Context, authorID int64, slug string, filename string, reader io.Reader, size int64) (*domain.ParseJob, error) {
	set, err := s.sets.EnsureAuthorOwns(ctx, authorID, slug)
	if err != nil {
		return nil, err
	}
	if err := s.sets.CheckVersionLimit(ctx, set.ID); err != nil {
		return nil, err
	}

	meta, err := worker.ParseZipFileName(filename)
	if err != nil {
		return nil, err
	}
	if meta.Slug != slug {
		return nil, fmt.Errorf("zip slug %s does not match set %s", meta.Slug, slug)
	}

	maxBytes := int64(s.maxZipMB) * 1024 * 1024
	if size > maxBytes {
		return nil, fmt.Errorf("zip exceeds %d MB limit", s.maxZipMB)
	}

	workDir := filepath.Join(s.tempDir, fmt.Sprintf("upload_%s_%d", slug, os.Getpid()))
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, err
	}
	zipPath := filepath.Join(workDir, "upload.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		return nil, err
	}
	n, err := io.Copy(f, reader)
	f.Close()
	if err != nil {
		return nil, err
	}
	if n > maxBytes {
		return nil, fmt.Errorf("zip exceeds %d MB limit", s.maxZipMB)
	}

	version := &domain.LevelSetVersion{
		SetID:         set.ID,
		Version:       meta.Version,
		ParseStatus:   domain.ParseStatusQueued,
		ParsePhase:    "queued",
		ParseProgress: 0,
		ParseMessage:  "等待解析…",
	}
	if err := s.versions.Create(ctx, version); err != nil {
		return nil, err
	}

	job := &domain.ParseJob{
		VersionID: version.ID,
		Status:    domain.JobStatusQueued,
		TempDir:   workDir,
	}
	if err := s.jobs.Create(ctx, job); err != nil {
		return nil, err
	}

	s.queue.Enqueue(job.ID)
	return job, nil
}

func (s *UploadService) GetJob(ctx context.Context, authorID int64, jobID int64) (*domain.ParseJob, error) {
	job, err := s.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}
	version, err := s.versions.GetByID(ctx, job.VersionID)
	if err != nil {
		return nil, err
	}
	set, err := s.sets.GetByID(ctx, version.SetID)
	if err != nil {
		return nil, err
	}
	if set.AuthorID != authorID {
		return nil, ErrForbidden
	}
	return job, nil
}
