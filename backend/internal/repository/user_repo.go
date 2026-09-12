package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO users (username, password_md5, role, must_change_password, display_name)
		VALUES (?, ?, ?, ?, ?)`,
		u.Username, u.PasswordMD5, u.Role, u.MustChangePassword, u.DisplayName)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	u.ID = id
	return nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_md5, role, must_change_password, display_name, avatar_cos_key, created_at, disabled_at
		FROM users WHERE username = ?`, username)
	return scanUser(row)
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, username, password_md5, role, must_change_password, display_name, avatar_cos_key, created_at, disabled_at
		FROM users WHERE id = ?`, id)
	return scanUser(row)
}

func (r *UserRepo) List(ctx context.Context) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, password_md5, role, must_change_password, display_name, avatar_cos_key, created_at, disabled_at
		FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.User
	for rows.Next() {
		u, err := scanUserRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *u)
	}
	return list, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordMD5 string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET password_md5 = ?, must_change_password = 0 WHERE id = ?`, passwordMD5, id)
	return err
}

func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET role = ?, display_name = ?, disabled_at = ? WHERE id = ?`,
		u.Role, u.DisplayName, nullableTime(u.DisabledAt), u.ID)
	return err
}

func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var n int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var mustChange int
	var createdAt string
	var disabledAt sql.NullString
	err := row.Scan(&u.ID, &u.Username, &u.PasswordMD5, &u.Role, &mustChange,
		&u.DisplayName, &u.AvatarCOSKey, &createdAt, &disabledAt)
	if err != nil {
		return nil, err
	}
	u.MustChangePassword = mustChange == 1
	u.CreatedAt = parseTime(createdAt)
	if disabledAt.Valid {
		t := parseTime(disabledAt.String)
		u.DisabledAt = &t
	}
	return &u, nil
}

func scanUserRow(rows *sql.Rows) (*domain.User, error) {
	var u domain.User
	var mustChange int
	var createdAt string
	var disabledAt sql.NullString
	err := rows.Scan(&u.ID, &u.Username, &u.PasswordMD5, &u.Role, &mustChange,
		&u.DisplayName, &u.AvatarCOSKey, &createdAt, &disabledAt)
	if err != nil {
		return nil, err
	}
	u.MustChangePassword = mustChange == 1
	u.CreatedAt = parseTime(createdAt)
	if disabledAt.Valid {
		t := parseTime(disabledAt.String)
		u.DisabledAt = &t
	}
	return &u, nil
}

func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Format("2006-01-02 15:04:05")
}
