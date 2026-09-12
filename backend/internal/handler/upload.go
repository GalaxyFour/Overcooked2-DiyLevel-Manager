package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/middleware"
	"github.com/gua248/Overcooked2-DiyLevel-Manager/internal/service"
)

type UploadHandler struct {
	upload *service.UploadService
}

func NewUploadHandler(upload *service.UploadService) *UploadHandler {
	return &UploadHandler{upload: upload}
}

// Upload godoc
// @Summary Upload level set zip
// @Tags author
// @Accept mpfd
// @Produce json
// @Param slug path string true "set slug"
// @Param file formData file true "zip file"
// @Router /api/v1/me/sets/{slug}/upload [post]
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	slug := chi.URLParam(r, "slug")

	if err := r.ParseMultipartForm(512 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	job, err := h.upload.Upload(r.Context(), u.ID, slug, header.Filename, file, header.Size)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, job)
}

// GetJob godoc
// @Summary Get parse job progress (poll every 2s)
// @Tags author
// @Produce json
// @Param id path int true "job id"
// @Router /api/v1/me/parse-jobs/{id} [get]
func (h *UploadHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	u := middleware.UserFromContext(r.Context())
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}
	job, err := h.upload.GetJob(r.Context(), u.ID, id)
	if err != nil {
		writeError(w, http.StatusForbidden, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, job)
}
