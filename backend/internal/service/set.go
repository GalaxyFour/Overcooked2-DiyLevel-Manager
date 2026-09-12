package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/repository"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/storage"
)

const (
	MaxSetsPerAuthor    = 20
	MaxVersionsPerSet   = 200
)

var (
	ErrSetLimitReached     = errors.New("author set limit reached (20)")
	ErrVersionLimitReached = errors.New("version limit reached (200)")
	ErrInvalidSlug         = errors.New("invalid slug")
	ErrSetNotFound         = errors.New("set not found")
	ErrForbidden           = errors.New("forbidden")
)

var slugRe = regexp.MustCompile(`^[a-z][a-z0-9_]{1,48}$`)

type SetService struct {
	sets     *repository.SetRepo
	versions *repository.VersionRepo
	entries  *repository.EntryRepo
	storage  *storage.COSStorage
}

func NewSetService(sets *repository.SetRepo, versions *repository.VersionRepo, entries *repository.EntryRepo, storage *storage.COSStorage) *SetService {
	return &SetService{sets: sets, versions: versions, entries: entries, storage: storage}
}

func (s *SetService) CreateSet(ctx context.Context, authorID int64, slug, name, nameZH string) (*domain.LevelSet, error) {
	slug = strings.TrimSpace(slug)
	if !slugRe.MatchString(slug) {
		return nil, ErrInvalidSlug
	}
	n, err := s.sets.CountByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}
	if n >= MaxSetsPerAuthor {
		return nil, ErrSetLimitReached
	}
	set := &domain.LevelSet{
		Slug:     slug,
		AuthorID: authorID,
		Name:     name,
		NameZH:   nameZH,
		Status:   domain.SetStatusActive,
	}
	if err := s.sets.Create(ctx, set); err != nil {
		return nil, err
	}
	return set, nil
}

func (s *SetService) ListMySets(ctx context.Context, authorID int64) ([]domain.LevelSet, error) {
	sets, err := s.sets.ListByAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}
	return s.enrichCovers(ctx, sets)
}

func (s *SetService) ListPublic(ctx context.Context) ([]domain.LevelSet, error) {
	sets, err := s.sets.ListPublic(ctx)
	if err != nil {
		return nil, err
	}
	return s.enrichCovers(ctx, sets)
}

func (s *SetService) ListAll(ctx context.Context, status string) ([]domain.LevelSet, error) {
	sets, err := s.sets.ListAll(ctx, status)
	if err != nil {
		return nil, err
	}
	return s.enrichCovers(ctx, sets)
}

func (s *SetService) GetBySlug(ctx context.Context, slug string) (*domain.LevelSet, []domain.LevelSetVersion, []domain.LevelEntry, error) {
	set, err := s.sets.GetBySlug(ctx, slug)
	if err != nil {
		return nil, nil, nil, ErrSetNotFound
	}
	versions, err := s.versions.ListBySet(ctx, set.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	var entries []domain.LevelEntry
	if len(versions) > 0 {
		for _, v := range versions {
			if v.IsLatest && v.ParseStatus == domain.ParseStatusDone {
				entries, err = s.entries.ListByVersion(ctx, v.ID)
				if err != nil {
					return nil, nil, nil, err
				}
				for i := range entries {
					if entries[i].ScreenshotCOSKey != "" {
						p, _ := s.storage.Presign(ctx, entries[i].ScreenshotCOSKey, domain.PresignKindImage)
						if p != nil {
							entries[i].ScreenshotURL = p.URL
						}
					}
				}
				break
			}
		}
	}
	sets, _ := s.enrichCovers(ctx, []domain.LevelSet{*set})
	if len(sets) > 0 {
		*set = sets[0]
	}
	return set, versions, entries, nil
}

func (s *SetService) UpdateSet(ctx context.Context, user *domain.User, slug string, name, nameZH string, status *domain.SetStatus) (*domain.LevelSet, error) {
	set, err := s.sets.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrSetNotFound
	}
	if user.Role == domain.RoleAuthor && set.AuthorID != user.ID {
		return nil, ErrForbidden
	}
	if name != "" {
		set.Name = name
	}
	if nameZH != "" {
		set.NameZH = nameZH
	}
	if status != nil {
		if user.Role == domain.RoleAuthor && *status != domain.SetStatusHidden && *status != domain.SetStatusActive {
			return nil, ErrForbidden
		}
		set.Status = *status
	}
	if err := s.sets.Update(ctx, set); err != nil {
		return nil, err
	}
	return set, nil
}

func (s *SetService) GetDownloadURL(ctx context.Context, slug, version string) (*domain.PresignResult, error) {
	set, err := s.sets.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrSetNotFound
	}
	var v *domain.LevelSetVersion
	if version == "" || version == "latest" {
		v, err = s.versions.GetLatest(ctx, set.ID)
	} else {
		v, err = s.versions.GetBySetAndVersion(ctx, set.ID, version)
	}
	if err != nil || v == nil || v.ZipCOSKey == "" {
		return nil, fmt.Errorf("version not available")
	}
	return s.storage.Presign(ctx, v.ZipCOSKey, domain.PresignKindPackage)
}

func (s *SetService) enrichCovers(ctx context.Context, sets []domain.LevelSet) ([]domain.LevelSet, error) {
	for i := range sets {
		if sets[i].CoverCOSKey != "" {
			p, err := s.storage.Presign(ctx, sets[i].CoverCOSKey, domain.PresignKindImage)
			if err == nil && p != nil {
				sets[i].CoverURL = p.URL
			}
		}
	}
	return sets, nil
}

func (s *SetService) EnsureAuthorOwns(ctx context.Context, userID int64, slug string) (*domain.LevelSet, error) {
	set, err := s.sets.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrSetNotFound
	}
	if set.AuthorID != userID {
		return nil, ErrForbidden
	}
	return set, nil
}

func (s *SetService) CheckVersionLimit(ctx context.Context, setID int64) error {
	n, err := s.versions.CountBySet(ctx, setID)
	if err != nil {
		return err
	}
	if n >= MaxVersionsPerSet {
		return ErrVersionLimitReached
	}
	return nil
}

func (s *SetService) UpdateStatus(ctx context.Context, setID int64, status domain.SetStatus) error {
	return s.sets.UpdateStatus(ctx, setID, status)
}

func (s *SetService) GetByID(ctx context.Context, id int64) (*domain.LevelSet, error) {
	return s.sets.GetByID(ctx, id)
}
