package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
)

type SetRepo struct {
	db *sql.DB
}

func NewSetRepo(db *sql.DB) *SetRepo {
	return &SetRepo{db: db}
}

func (r *SetRepo) Create(ctx context.Context, s *domain.LevelSet) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO level_sets (slug, author_id, name, name_zh, status)
		VALUES (?, ?, ?, ?, ?)`,
		s.Slug, s.AuthorID, s.Name, s.NameZH, s.Status)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	s.ID = id
	return nil
}

func (r *SetRepo) GetBySlug(ctx context.Context, slug string) (*domain.LevelSet, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT ls.id, ls.slug, ls.author_id, ls.name, ls.name_zh, ls.status, ls.cover_cos_key, ls.created_at,
		       COALESCE(u.display_name, u.username, '')
		FROM level_sets ls
		LEFT JOIN users u ON u.id = ls.author_id
		WHERE ls.slug = ?`, slug)
	return scanSet(row)
}

func (r *SetRepo) GetByID(ctx context.Context, id int64) (*domain.LevelSet, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT ls.id, ls.slug, ls.author_id, ls.name, ls.name_zh, ls.status, ls.cover_cos_key, ls.created_at,
		       COALESCE(u.display_name, u.username, '')
		FROM level_sets ls
		LEFT JOIN users u ON u.id = ls.author_id
		WHERE ls.id = ?`, id)
	return scanSet(row)
}

func (r *SetRepo) ListByAuthor(ctx context.Context, authorID int64) ([]domain.LevelSet, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ls.id, ls.slug, ls.author_id, ls.name, ls.name_zh, ls.status, ls.cover_cos_key, ls.created_at,
		       COALESCE(u.display_name, u.username, ''),
		       COALESCE(v.version, ''), COALESCE((SELECT COUNT(*) FROM level_entries le WHERE le.version_id = v.id), 0)
		FROM level_sets ls
		LEFT JOIN users u ON u.id = ls.author_id
		LEFT JOIN level_set_versions v ON v.set_id = ls.id AND v.is_latest = 1
		WHERE ls.author_id = ?
		ORDER BY ls.created_at DESC`, authorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSetRows(rows)
}

func (r *SetRepo) ListPublic(ctx context.Context) ([]domain.LevelSet, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ls.id, ls.slug, ls.author_id, ls.name, ls.name_zh, ls.status, ls.cover_cos_key, ls.created_at,
		       COALESCE(u.display_name, u.username, ''),
		       COALESCE(v.version, ''), COALESCE((SELECT COUNT(*) FROM level_entries le WHERE le.version_id = v.id), 0)
		FROM level_sets ls
		LEFT JOIN users u ON u.id = ls.author_id
		LEFT JOIN level_set_versions v ON v.set_id = ls.id AND v.is_latest = 1 AND v.parse_status = 'done'
		WHERE ls.status = 'active'
		ORDER BY ls.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSetRows(rows)
}

func (r *SetRepo) ListAll(ctx context.Context, status string) ([]domain.LevelSet, error) {
	q := `
		SELECT ls.id, ls.slug, ls.author_id, ls.name, ls.name_zh, ls.status, ls.cover_cos_key, ls.created_at,
		       COALESCE(u.display_name, u.username, ''),
		       COALESCE(v.version, ''), COALESCE((SELECT COUNT(*) FROM level_entries le WHERE le.version_id = v.id), 0)
		FROM level_sets ls
		LEFT JOIN users u ON u.id = ls.author_id
		LEFT JOIN level_set_versions v ON v.set_id = ls.id AND v.is_latest = 1
		WHERE 1=1`
	args := []interface{}{}
	if status != "" {
		q += ` AND ls.status = ?`
		args = append(args, status)
	}
	q += ` ORDER BY ls.created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSetRows(rows)
}

func (r *SetRepo) CountByAuthor(ctx context.Context, authorID int64) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM level_sets WHERE author_id = ?`, authorID).Scan(&n)
	return n, err
}

func (r *SetRepo) Update(ctx context.Context, s *domain.LevelSet) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE level_sets SET name = ?, name_zh = ?, status = ?, cover_cos_key = ? WHERE id = ?`,
		s.Name, s.NameZH, s.Status, s.CoverCOSKey, s.ID)
	return err
}

func (r *SetRepo) UpdateStatus(ctx context.Context, id int64, status domain.SetStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE level_sets SET status = ? WHERE id = ?`, status, id)
	return err
}

func scanSet(row *sql.Row) (*domain.LevelSet, error) {
	var s domain.LevelSet
	var createdAt string
	err := row.Scan(&s.ID, &s.Slug, &s.AuthorID, &s.Name, &s.NameZH, &s.Status,
		&s.CoverCOSKey, &createdAt, &s.AuthorName)
	if err != nil {
		return nil, err
	}
	s.CreatedAt = parseTime(createdAt)
	return &s, nil
}

func scanSetRows(rows *sql.Rows) ([]domain.LevelSet, error) {
	var list []domain.LevelSet
	for rows.Next() {
		var s domain.LevelSet
		var createdAt string
		if err := rows.Scan(&s.ID, &s.Slug, &s.AuthorID, &s.Name, &s.NameZH, &s.Status,
			&s.CoverCOSKey, &createdAt, &s.AuthorName, &s.LatestVersion, &s.LevelCount); err != nil {
			return nil, err
		}
		s.CreatedAt = parseTime(createdAt)
		list = append(list, s)
	}
	return list, nil
}

type VersionRepo struct {
	db *sql.DB
}

func NewVersionRepo(db *sql.DB) *VersionRepo {
	return &VersionRepo{db: db}
}

func (r *VersionRepo) Create(ctx context.Context, v *domain.LevelSetVersion) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO level_set_versions (set_id, version, uid, zip_cos_key, file_count, has_runtime, has_common_w2,
			parse_status, parse_phase, parse_progress, parse_message, is_latest)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.SetID, v.Version, v.UID, v.ZipCOSKey, v.FileCount, boolToInt(v.HasRuntime), boolToInt(v.HasCommonW2),
		v.ParseStatus, v.ParsePhase, v.ParseProgress, v.ParseMessage, boolToInt(v.IsLatest))
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	v.ID = id
	return nil
}

func (r *VersionRepo) GetByID(ctx context.Context, id int64) (*domain.LevelSetVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, set_id, version, uid, zip_cos_key, file_count, has_runtime, has_common_w2,
			parse_status, parse_phase, parse_progress, parse_message, is_latest, created_at
		FROM level_set_versions WHERE id = ?`, id)
	return scanVersion(row)
}

func (r *VersionRepo) GetBySetAndVersion(ctx context.Context, setID int64, version string) (*domain.LevelSetVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, set_id, version, uid, zip_cos_key, file_count, has_runtime, has_common_w2,
			parse_status, parse_phase, parse_progress, parse_message, is_latest, created_at
		FROM level_set_versions WHERE set_id = ? AND version = ?`, setID, version)
	return scanVersion(row)
}

func (r *VersionRepo) GetLatest(ctx context.Context, setID int64) (*domain.LevelSetVersion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, set_id, version, uid, zip_cos_key, file_count, has_runtime, has_common_w2,
			parse_status, parse_phase, parse_progress, parse_message, is_latest, created_at
		FROM level_set_versions WHERE set_id = ? AND is_latest = 1`, setID)
	return scanVersion(row)
}

func (r *VersionRepo) ListBySet(ctx context.Context, setID int64) ([]domain.LevelSetVersion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, set_id, version, uid, zip_cos_key, file_count, has_runtime, has_common_w2,
			parse_status, parse_phase, parse_progress, parse_message, is_latest, created_at
		FROM level_set_versions WHERE set_id = ? ORDER BY created_at DESC`, setID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.LevelSetVersion
	for rows.Next() {
		v, err := scanVersionRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *v)
	}
	return list, nil
}

func (r *VersionRepo) CountBySet(ctx context.Context, setID int64) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM level_set_versions WHERE set_id = ?`, setID).Scan(&n)
	return n, err
}

func (r *VersionRepo) UpdateProgress(ctx context.Context, id int64, status domain.ParseStatus, phase string, progress int, message string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE level_set_versions SET parse_status = ?, parse_phase = ?, parse_progress = ?, parse_message = ?
		WHERE id = ?`, status, phase, progress, message, id)
	return err
}

func (r *VersionRepo) Finalize(ctx context.Context, v *domain.LevelSetVersion) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE level_set_versions SET is_latest = 0 WHERE set_id = ?`, v.SetID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE level_set_versions SET uid = ?, zip_cos_key = ?, file_count = ?, has_runtime = ?, has_common_w2 = ?,
			parse_status = 'done', parse_phase = 'done', parse_progress = 100, parse_message = ?, is_latest = 1
		WHERE id = ?`,
		v.UID, v.ZipCOSKey, v.FileCount, boolToInt(v.HasRuntime), boolToInt(v.HasCommonW2), v.ParseMessage, v.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *VersionRepo) MarkFailed(ctx context.Context, id int64, message string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE level_set_versions SET parse_status = 'failed', parse_phase = 'failed', parse_message = ?
		WHERE id = ?`, message, id)
	return err
}

func scanVersion(row *sql.Row) (*domain.LevelSetVersion, error) {
	var v domain.LevelSetVersion
	var hasRuntime, hasCommonW2, isLatest int
	var createdAt string
	err := row.Scan(&v.ID, &v.SetID, &v.Version, &v.UID, &v.ZipCOSKey, &v.FileCount,
		&hasRuntime, &hasCommonW2, &v.ParseStatus, &v.ParsePhase, &v.ParseProgress, &v.ParseMessage,
		&isLatest, &createdAt)
	if err != nil {
		return nil, err
	}
	v.HasRuntime = hasRuntime == 1
	v.HasCommonW2 = hasCommonW2 == 1
	v.IsLatest = isLatest == 1
	v.CreatedAt = parseTime(createdAt)
	return &v, nil
}

func scanVersionRow(rows *sql.Rows) (*domain.LevelSetVersion, error) {
	var v domain.LevelSetVersion
	var hasRuntime, hasCommonW2, isLatest int
	var createdAt string
	err := rows.Scan(&v.ID, &v.SetID, &v.Version, &v.UID, &v.ZipCOSKey, &v.FileCount,
		&hasRuntime, &hasCommonW2, &v.ParseStatus, &v.ParsePhase, &v.ParseProgress, &v.ParseMessage,
		&isLatest, &createdAt)
	if err != nil {
		return nil, err
	}
	v.HasRuntime = hasRuntime == 1
	v.HasCommonW2 = hasCommonW2 == 1
	v.IsLatest = isLatest == 1
	v.CreatedAt = parseTime(createdAt)
	return &v, nil
}

type EntryRepo struct {
	db *sql.DB
}

func NewEntryRepo(db *sql.DB) *EntryRepo {
	return &EntryRepo{db: db}
}

func (r *EntryRepo) CreateBatch(ctx context.Context, entries []domain.LevelEntry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO level_entries (version_id, level_id, level_name, level_name_zh, scene_name, screenshot_cos_key, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, e := range entries {
		if _, err := stmt.ExecContext(ctx, e.VersionID, e.LevelID, e.LevelName, e.LevelNameZH, e.SceneName, e.ScreenshotCOSKey, e.SortOrder); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *EntryRepo) ListByVersion(ctx context.Context, versionID int64) ([]domain.LevelEntry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, version_id, level_id, level_name, level_name_zh, scene_name, screenshot_cos_key, sort_order
		FROM level_entries WHERE version_id = ? ORDER BY sort_order`, versionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.LevelEntry
	for rows.Next() {
		var e domain.LevelEntry
		if err := rows.Scan(&e.ID, &e.VersionID, &e.LevelID, &e.LevelName, &e.LevelNameZH, &e.SceneName, &e.ScreenshotCOSKey, &e.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

type JobRepo struct {
	db *sql.DB
}

func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(ctx context.Context, j *domain.ParseJob) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO parse_jobs (version_id, status, temp_dir) VALUES (?, ?, ?)`,
		j.VersionID, j.Status, j.TempDir)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	j.ID = id
	return nil
}

func (r *JobRepo) GetByID(ctx context.Context, id int64) (*domain.ParseJob, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT j.id, j.version_id, j.status, j.temp_dir, j.started_at, j.finished_at, j.error,
			v.parse_phase, v.parse_progress, v.parse_message
		FROM parse_jobs j
		JOIN level_set_versions v ON v.id = j.version_id
		WHERE j.id = ?`, id)
	var j domain.ParseJob
	var startedAt, finishedAt sql.NullString
	err := row.Scan(&j.ID, &j.VersionID, &j.Status, &j.TempDir, &startedAt, &finishedAt, &j.Error,
		&j.Phase, &j.Progress, &j.Message)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		t := parseTime(startedAt.String)
		j.StartedAt = &t
	}
	if finishedAt.Valid {
		t := parseTime(finishedAt.String)
		j.FinishedAt = &t
	}
	return &j, nil
}

func (r *JobRepo) GetNextQueued(ctx context.Context) (*domain.ParseJob, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, version_id, status, temp_dir, started_at, finished_at, error
		FROM parse_jobs WHERE status = 'queued' ORDER BY id LIMIT 1`)
	var j domain.ParseJob
	var startedAt, finishedAt sql.NullString
	err := row.Scan(&j.ID, &j.VersionID, &j.Status, &j.TempDir, &startedAt, &finishedAt, &j.Error)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *JobRepo) UpdateStatus(ctx context.Context, id int64, status domain.JobStatus, errMsg string) error {
	now := timeNow()
	if status == domain.JobStatusRunning {
		_, err := r.db.ExecContext(ctx, `UPDATE parse_jobs SET status = ?, started_at = ? WHERE id = ?`, status, now, id)
		return err
	}
	if status == domain.JobStatusDone || status == domain.JobStatusFailed {
		_, err := r.db.ExecContext(ctx, `UPDATE parse_jobs SET status = ?, finished_at = ?, error = ? WHERE id = ?`, status, now, errMsg, id)
		return err
	}
	_, err := r.db.ExecContext(ctx, `UPDATE parse_jobs SET status = ? WHERE id = ?`, status, id)
	return err
}

func (r *JobRepo) GetByVersionID(ctx context.Context, versionID int64) (*domain.ParseJob, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT j.id, j.version_id, j.status, j.temp_dir, j.started_at, j.finished_at, j.error,
			v.parse_phase, v.parse_progress, v.parse_message
		FROM parse_jobs j
		JOIN level_set_versions v ON v.id = j.version_id
		WHERE j.version_id = ? ORDER BY j.id DESC LIMIT 1`, versionID)
	var j domain.ParseJob
	var startedAt, finishedAt sql.NullString
	err := row.Scan(&j.ID, &j.VersionID, &j.Status, &j.TempDir, &startedAt, &finishedAt, &j.Error,
		&j.Phase, &j.Progress, &j.Message)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		t := parseTime(startedAt.String)
		j.StartedAt = &t
	}
	if finishedAt.Valid {
		t := parseTime(finishedAt.String)
		j.FinishedAt = &t
	}
	return &j, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func timeNow() string {
	return time.Now().UTC().Format("2006-01-02 15:04:05")
}
