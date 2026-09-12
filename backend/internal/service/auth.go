package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/config"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDisabled       = errors.New("user disabled")
	ErrMustChangePassword = errors.New("must change password")
)

type AuthService struct {
	users  *repository.UserRepo
	secret string
}

func NewAuthService(users *repository.UserRepo, cfg *config.ServerConfig) *AuthService {
	return &AuthService{users: users, secret: cfg.JWTSecret}
}

func HashPasswordMD5(password string) string {
	h := md5.Sum([]byte(password))
	return hex.EncodeToString(h[:])
}

func (s *AuthService) Login(ctx context.Context, username, password string) (string, *domain.User, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}
	if u.DisabledAt != nil {
		return "", nil, ErrUserDisabled
	}
	if u.PasswordMD5 != HashPasswordMD5(password) {
		return "", nil, ErrInvalidCredentials
	}
	token, err := s.issueToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, newPassword string) error {
	return s.users.UpdatePassword(ctx, userID, HashPasswordMD5(newPassword))
}

func (s *AuthService) ParseToken(tokenStr string) (*domain.User, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(s.secret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidCredentials
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	userID, ok := claims["uid"].(float64)
	if !ok {
		return nil, ErrInvalidCredentials
	}
	u, err := s.users.GetByID(context.Background(), int64(userID))
	if err != nil {
		return nil, err
	}
	if u.DisabledAt != nil {
		return nil, ErrUserDisabled
	}
	return u, nil
}

func (s *AuthService) issueToken(u *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"uid":  u.ID,
		"role": string(u.Role),
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *AuthService) BootstrapSuperAdmin(ctx context.Context, username, password string) error {
	n, err := s.users.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	u := &domain.User{
		Username:           username,
		PasswordMD5:        HashPasswordMD5(password),
		Role:               domain.RoleSuperAdmin,
		MustChangePassword: false,
		DisplayName:        "Super Admin",
	}
	return s.users.Create(ctx, u)
}
