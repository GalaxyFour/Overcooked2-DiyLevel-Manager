package service

import (
	"context"
	"errors"
	"time"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/repository"
)

var ErrUserExists = errors.New("username already exists")

type UserService struct {
	users *repository.UserRepo
}

func NewUserService(users *repository.UserRepo) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Create(ctx context.Context, username, password, displayName string, role domain.Role) (*domain.User, error) {
	if _, err := s.users.GetByUsername(ctx, username); err == nil {
		return nil, ErrUserExists
	}
	u := &domain.User{
		Username:           username,
		PasswordMD5:        HashPasswordMD5(password),
		Role:               role,
		MustChangePassword: true,
		DisplayName:        displayName,
	}
	if displayName == "" {
		u.DisplayName = username
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	return s.users.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id int64, role domain.Role, displayName string, disabled bool) error {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return err
	}
	u.Role = role
	u.DisplayName = displayName
	if disabled {
		now := time.Now().UTC()
		u.DisabledAt = &now
	} else {
		u.DisabledAt = nil
	}
	return s.users.Update(ctx, u)
}
