package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/middleware"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/service"
)

type AuthHandler struct {
	auth       *service.AuthService
	cookieName string
}

func NewAuthHandler(auth *service.AuthService, cookieName string) *AuthHandler {
	return &AuthHandler{auth: auth, cookieName: cookieName}
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type changePasswordReq struct {
	NewPassword string `json:"newPassword"`
}

type userResp struct {
	ID                 int64  `json:"id"`
	Username           string `json:"username"`
	Role               string `json:"role"`
	MustChangePassword bool   `json:"mustChangePassword"`
	DisplayName        string `json:"displayName"`
}

func toUserResp(u *domain.User) userResp {
	return userResp{
		ID:                 u.ID,
		Username:           u.Username,
		Role:               string(u.Role),
		MustChangePassword: u.MustChangePassword,
		DisplayName:        u.DisplayName,
	}
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body loginReq true "credentials"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	token, user, err := h.auth.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
	})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":              token,
		"user":               toUserResp(user),
		"mustChangePassword": user.MustChangePassword,
	})
}

// Me godoc
// @Summary Current user
// @Tags auth
// @Produce json
// @Success 200 {object} userResp
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, toUserResp(u))
}

// ChangePassword godoc
// @Summary Change password
// @Tags auth
// @Accept json
// @Produce json
// @Router /api/v1/auth/change-password [post]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req changePasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.NewPassword) < 6 {
		writeError(w, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}
	if err := h.auth.ChangePassword(r.Context(), u.ID, req.NewPassword); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
