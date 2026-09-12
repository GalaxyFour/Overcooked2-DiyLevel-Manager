package storage

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/config"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
)

const (
	packageTTL = 6 * time.Hour
	imageTTL   = 48 * time.Hour
)

type COSStorage struct {
	client    *cos.Client
	secretID  string
	secretKey string
	bucket    string
	baseURL   string
	db        *sql.DB
	localDir  string
	mu        sync.Mutex
}

func NewCOSStorage(cfg *config.COSConfig, db *sql.DB, localDir string) (*COSStorage, error) {
	s := &COSStorage{
		db:        db,
		localDir:  localDir,
		bucket:    cfg.Bucket,
		baseURL:   cfg.BaseURL,
		secretID:  cfg.SecretID,
		secretKey: cfg.SecretKey,
	}
	if !cfg.Enabled() {
		if err := os.MkdirAll(localDir, 0o755); err != nil {
			return nil, err
		}
		return s, nil
	}
	u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region))
	b := &cos.BaseURL{BucketURL: u}
	s.client = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})
	return s, nil
}

func (s *COSStorage) Upload(ctx context.Context, key string, reader io.Reader) error {
	if s.client == nil {
		path := filepath.Join(s.localDir, key)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, reader)
		return err
	}
	_, err := s.client.Object.Put(ctx, key, reader, nil)
	return err
}

func (s *COSStorage) UploadFile(ctx context.Context, key, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return s.Upload(ctx, key, f)
}

func (s *COSStorage) Presign(ctx context.Context, key string, kind domain.PresignKind) (*domain.PresignResult, error) {
	if key == "" {
		return nil, fmt.Errorf("empty cos key")
	}
	ttl := packageTTL
	if kind == domain.PresignKindImage {
		ttl = imageTTL
	}

	cached, err := s.getCached(key)
	if err == nil && cached != nil {
		refreshBefore := cached.ExpiresAt.Add(-time.Duration(float64(ttl) * 0.1))
		if time.Now().Before(refreshBefore) {
			return cached, nil
		}
	}

	var result *domain.PresignResult
	if s.client == nil {
		localPath := filepath.Join(s.localDir, key)
		if _, err := os.Stat(localPath); err != nil {
			return nil, fmt.Errorf("local file not found: %s", key)
		}
		result = &domain.PresignResult{
			URL:       "/api/v1/local-files/" + url.PathEscape(key),
			ExpiresAt: time.Now().Add(ttl),
		}
	} else {
		presigned, err := s.client.Object.GetPresignedURL(ctx, http.MethodGet, key, s.secretID, s.secretKey, ttl, nil)
		if err != nil {
			return nil, err
		}
		result = &domain.PresignResult{
			URL:       presigned.String(),
			ExpiresAt: time.Now().Add(ttl),
		}
	}

	_ = s.saveCache(key, result, kind)
	return result, nil
}

func (s *COSStorage) ServeLocal(key string) (string, error) {
	path := filepath.Join(s.localDir, key)
	clean := filepath.Clean(path)
	if !strings.HasPrefix(clean, filepath.Clean(s.localDir)) {
		return "", fmt.Errorf("invalid path")
	}
	if _, err := os.Stat(clean); err != nil {
		return "", err
	}
	return clean, nil
}

func (s *COSStorage) getCached(key string) (*domain.PresignResult, error) {
	row := s.db.QueryRow(`SELECT url, expires_at FROM presign_cache WHERE cos_key = ?`, key)
	var u string
	var exp string
	if err := row.Scan(&u, &exp); err != nil {
		return nil, err
	}
	t, _ := time.Parse("2006-01-02 15:04:05", exp)
	if t.IsZero() {
		t, _ = time.Parse(time.RFC3339, exp)
	}
	return &domain.PresignResult{URL: u, ExpiresAt: t}, nil
}

func (s *COSStorage) saveCache(key string, r *domain.PresignResult, kind domain.PresignKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`
		INSERT INTO presign_cache (cos_key, url, expires_at, kind) VALUES (?, ?, ?, ?)
		ON CONFLICT(cos_key) DO UPDATE SET url = excluded.url, expires_at = excluded.expires_at, kind = excluded.kind`,
		key, r.URL, r.ExpiresAt.UTC().Format("2006-01-02 15:04:05"), kind)
	return err
}

func PackageKey(authorID int64, slug, version, filename string) string {
	return fmt.Sprintf("packages/%d/%s/%s/%s", authorID, slug, version, filename)
}

func ScreenshotKey(authorID int64, slug, version, levelID string) string {
	return fmt.Sprintf("screenshots/%d/%s/%s/%s.png", authorID, slug, version, levelID)
}

func CoverKey(slug string) string {
	return fmt.Sprintf("covers/%s/latest.png", slug)
}
