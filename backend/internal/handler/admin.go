package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/service"
)

type AdminHandler struct {
	sets  *service.SetService
	users *service.UserService
}

func NewAdminHandler(sets *service.SetService, users *service.UserService) *AdminHandler {
	return &AdminHandler{sets: sets, users: users}
}

type createUserReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type patchUserReq struct {
	Role        string `json:"role"`
	DisplayName string `json:"displayName"`
	Disabled    bool   `json:"disabled"`
}

type patchStatusReq struct {
	Status string `json:"status"`
}

// AdminListSets godoc
// @Summary Admin list all sets
// @Tags admin
// @Router /api/v1/admin/sets [get]
func (h *AdminHandler) AdminListSets(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	sets, err := h.sets.ListAll(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sets)
}

// AdminPatchSetStatus godoc
// @Summary Admin update set status
// @Tags admin
// @Router /api/v1/admin/sets/{id}/status [patch]
func (h *AdminHandler) AdminPatchSetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req patchStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.sets.UpdateStatus(r.Context(), id, domain.SetStatus(req.Status)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListUsers godoc
// @Summary Super admin list users
// @Tags super
// @Router /api/v1/super/users [get]
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// CreateUser godoc
// @Summary Super admin create user
// @Tags super
// @Router /api/v1/super/users [post]
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Role == "" {
		req.Role = string(domain.RoleAuthor)
	}
	u, err := h.users.Create(r.Context(), req.Username, req.Password, req.DisplayName, domain.Role(req.Role))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

// PatchUser godoc
// @Summary Super admin update user
// @Tags super
// @Router /api/v1/super/users/{id} [patch]
func (h *AdminHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req patchUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.users.Update(r.Context(), id, domain.Role(req.Role), req.DisplayName, req.Disabled); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListInvalidSets godoc
// @Summary Super admin list invalid sets
// @Tags super
// @Router /api/v1/super/invalid-sets [get]
func (h *AdminHandler) ListInvalidSets(w http.ResponseWriter, r *http.Request) {
	sets, err := h.sets.ListAll(r.Context(), string(domain.SetStatusInvalid))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sets)
}

// InvalidateSet godoc
// @Summary Super admin mark set invalid
// @Tags super
// @Router /api/v1/super/sets/{id}/invalidate [post]
func (h *AdminHandler) InvalidateSet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.sets.UpdateStatus(r.Context(), id, domain.SetStatusInvalid); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
