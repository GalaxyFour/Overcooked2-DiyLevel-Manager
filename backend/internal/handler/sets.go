package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/domain"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/middleware"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/service"
)

type SetsHandler struct {
	sets *service.SetService
}

func NewSetsHandler(sets *service.SetService) *SetsHandler {
	return &SetsHandler{sets: sets}
}

// ListPublic godoc
// @Summary List all published level sets (latest version)
// @Tags public
// @Produce json
// @Success 200 {array} domain.LevelSet
// @Router /api/v1/sets [get]
func (h *SetsHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	sets, err := h.sets.ListPublic(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sets == nil {
		sets = []domain.LevelSet{}
	}
	writeJSON(w, http.StatusOK, sets)
}

// GetBySlug godoc
// @Summary Get level set detail
// @Tags public
// @Produce json
// @Param slug path string true "set slug"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sets/{slug} [get]
func (h *SetsHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	set, versions, entries, err := h.sets.GetBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"set":      set,
		"versions": versions,
		"levels":   entries,
	})
}

// DownloadLatest godoc
// @Summary Download latest version presigned URL
// @Tags public
// @Produce json
// @Param slug path string true "set slug"
// @Success 200 {object} domain.PresignResult
// @Router /api/v1/sets/{slug}/latest/download [get]
func (h *SetsHandler) DownloadLatest(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	result, err := h.sets.GetDownloadURL(r.Context(), slug, "")
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// DownloadVersion godoc
// @Summary Download specific version presigned URL
// @Tags public
// @Produce json
// @Param slug path string true "set slug"
// @Param version path string true "version"
// @Success 200 {object} domain.PresignResult
// @Router /api/v1/sets/{slug}/versions/{version}/download [get]
func (h *SetsHandler) DownloadVersion(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	version := chi.URLParam(r, "version")
	result, err := h.sets.GetDownloadURL(r.Context(), slug, version)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// BundleDownload godoc
// @Summary Bundle download alias for latest zip
// @Tags public
// @Produce json
// @Param slug path string true "set slug"
// @Success 200 {object} domain.PresignResult
// @Router /api/v1/sets/{slug}/bundle [get]
func (h *SetsHandler) BundleDownload(w http.ResponseWriter, r *http.Request) {
	h.DownloadLatest(w, r)
}

type createSetReq struct {
	Slug   string `json:"slug"`
	Name   string `json:"name"`
	NameZH string `json:"nameZh"`
}

type patchSetReq struct {
	Name   string `json:"name"`
	NameZH string `json:"nameZh"`
	Status string `json:"status"`
}

// ListMy godoc
// @Summary List author's level sets
// @Tags author
// @Produce json
// @Router /api/v1/me/sets [get]
func (h *SetsHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	sets, err := h.sets.ListMySets(r.Context(), u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if sets == nil {
		sets = []domain.LevelSet{}
	}
	writeJSON(w, http.StatusOK, sets)
}

// CreateSet godoc
// @Summary Create a new level set
// @Tags author
// @Accept json
// @Produce json
// @Router /api/v1/me/sets [post]
func (h *SetsHandler) CreateSet(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	var req createSetReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	set, err := h.sets.CreateSet(r.Context(), u.ID, req.Slug, req.Name, req.NameZH)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, set)
}

// PatchSet godoc
// @Summary Update level set metadata
// @Tags author
// @Accept json
// @Produce json
// @Param slug path string true "set slug"
// @Router /api/v1/me/sets/{slug} [patch]
func (h *SetsHandler) PatchSet(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	slug := chi.URLParam(r, "slug")
	var req patchSetReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var status *domain.SetStatus
	if req.Status != "" {
		s := domain.SetStatus(req.Status)
		status = &s
	}
	set, err := h.sets.UpdateSet(r.Context(), u, slug, req.Name, req.NameZH, status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, set)
}
