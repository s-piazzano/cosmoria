package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/s-piazzano/cosmoria/internal/auth"
)

// @Summary Download file content (local backend)
// @Description Stream file bytes directly for local storage. For S3, use the presigned URL from GET.
// @Tags Files
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param fid path string true "File ID"
// @Success 200 {file} binary
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /api/projects/{pid}/tenants/{tid}/files/{fid}/download [get]
func (h *FilesHandler) Download(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := claims.ProjectID
	tenantID := r.PathValue("tid")
	fileID := r.PathValue("fid")

	// Get file metadata (validates project+tenant ownership)
	f, err := h.Service.GetMeta(r.Context(), projectID, tenantID, fileID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}

	basePath := h.Service.StoragePath()
	fullPath := filepath.Join(basePath, f.S3Key)

	// Prevent path traversal
	if !isPathSafe(basePath, fullPath) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}

	file, err := os.Open(fullPath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "file not found"})
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+f.Filename+"\"")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, file)
}

func isPathSafe(base, candidate string) bool {
	absBase, err := filepath.Abs(base)
	if err != nil {
		return false
	}
	absCand, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absBase, absCand)
	if err != nil {
		return false
	}
	return rel == filepath.Clean(rel)
}
