package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/service"
)

type contextKey string

const UserContextKey contextKey = "user"

func Auth(auth *service.AuthService, cookieName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if c, err := r.Cookie(cookieName); err == nil {
				token = c.Value
			}
			if token == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}
			if token != "" {
				u, err := auth.ParseToken(token)
				if err == nil {
					ctx := context.WithValue(r.Context(), UserContextKey, u)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if UserFromContext(r.Context()) == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireRoles(roles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := UserFromContext(r.Context())
			if u == nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			for _, role := range roles {
				if u.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			if u.Role == domain.RoleSuperAdmin {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		})
	}
}

func UserFromContext(ctx context.Context) *domain.User {
	u, _ := ctx.Value(UserContextKey).(*domain.User)
	return u
}
