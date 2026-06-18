package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/realtime"
	"github.com/s-piazzano/cosmoria/internal/storage"
)

type FilesHandler struct {
	Service   *storage.Service
	Publisher *realtime.Publisher
}

type fileResponse struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"project_id"`
	TenantID     *string `json:"tenant_id"`
	Filename     string  `json:"filename"`
	Size         int64   `json:"size"`
	MimeType     string  `json:"mime_type"`
	UploadedBy   string  `json:"uploaded_by"`
	CreatedAt    string  `json:"created_at"`
	PresignedURL string  `json:"presigned_url,omitempty"`
}

func toFileResponse(f *storage.File) fileResponse {
	return fileResponse{
		ID:         f.ID,
		ProjectID:  f.ProjectID,
		TenantID:   f.TenantID,
		Filename:   f.Filename,
		Size:       f.Size,
		MimeType:   f.MimeType,
		UploadedBy: f.UploadedBy,
		CreatedAt:  f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Upload handles multipart file upload.
// @Summary Upload a file
// @Description Upload a file to S3-compatible storage and store metadata. Tenant-scoped.
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param file formData file true "File to upload"
// @Success 201 {object} fileResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/projects/{pid}/tenants/{tid}/files [post]
func (h *FilesHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := claims.ProjectID
	tenantID := strPtr(r.PathValue("tid"))

	maxSize := h.Service.MaxUploadSize()
	if maxSize > 0 {
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	}

	info, err := storage.ParseUpload(r, maxSize)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := storage.ValidateMimeType(info.MimeType); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	f, err := h.Service.Upload(r.Context(), projectID, tenantID, info.Filename, info.MimeType, info.Reader, info.Size, claims.UserID)
	if err != nil {
		slog.Error("file upload failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if h.Publisher != nil {
		h.Publisher.Publish(&realtime.Event{
			ProjectID:  projectID,
			TenantID:   safeStr(tenantID),
			Resource:   "files",
			Action:     "create",
			ResourceID: f.ID,
		})
	}

	writeJSON(w, http.StatusCreated, toFileResponse(f))
}

// List returns paginated files for a tenant.
// @Summary List files
// @Description List all files for a tenant with cursor-based pagination.
// @Tags Files
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param cursor query string false "Pagination cursor"
// @Param limit query int false "Max results (1-100)"
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/projects/{pid}/tenants/{tid}/files [get]
func (h *FilesHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := claims.ProjectID
	tenantID := strPtr(r.PathValue("tid"))
	cursor := r.URL.Query().Get("cursor")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	files, nextCursor, err := h.Service.List(r.Context(), projectID, tenantID, cursor, limit)
	if err != nil {
		slog.Error("file list failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	resp := make([]fileResponse, 0, len(files))
	for _, f := range files {
		resp = append(resp, toFileResponse(&f))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"files":       resp,
		"next_cursor": nextCursor,
	})
}

// Get returns a single file with a presigned download URL.
// @Summary Get file
// @Description Get file metadata and a presigned download URL.
// @Tags Files
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param fid path string true "File ID"
// @Success 200 {object} fileResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/projects/{pid}/tenants/{tid}/files/{fid} [get]
func (h *FilesHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := claims.ProjectID
	tenantID := strPtr(r.PathValue("tid"))
	fileID := r.PathValue("fid")

	f, err := h.Service.GetByID(r.Context(), projectID, tenantID, fileID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}

	resp := toFileResponse(&f.File)
	resp.PresignedURL = f.PresignedURL
	writeJSON(w, http.StatusOK, resp)
}

// Delete removes a file from storage and DB.
// @Summary Delete a file
// @Description Delete a file from S3 storage and remove its metadata.
// @Tags Files
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param fid path string true "File ID"
// @Success 204 "No Content"
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/projects/{pid}/tenants/{tid}/files/{fid} [delete]
func (h *FilesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := claims.ProjectID
	tenantID := strPtr(r.PathValue("tid"))
	fileID := r.PathValue("fid")

	if err := h.Service.Delete(r.Context(), projectID, tenantID, fileID); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}

	if h.Publisher != nil {
		h.Publisher.Publish(&realtime.Event{
			ProjectID:  projectID,
			TenantID:   safeStr(tenantID),
			Resource:   "files",
			Action:     "delete",
			ResourceID: fileID,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}
