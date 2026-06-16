package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/adminauth"
)

type ApiKeysHandler struct {
	Service *auth.ApiKeyService
}

type createApiKeyRequest struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

// @Summary Create API key
// @Security AdminBearerAuth
// @Description Create a new API key for a project, attached to a user. super_admin only.
// @Tags Admin
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param body body createApiKeyRequest true "User ID and key name"
// @Success 201 {object} auth.CreateApiKeyResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/api-keys [post]
func (h *ApiKeysHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if claims.Role != "super_admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	projectID := r.PathValue("pid")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_project_id"})
		return
	}

	var req createApiKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.UserID == "" || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id_and_name_required"})
		return
	}

	result, err := h.Service.Create(r.Context(), projectID, req.UserID, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// @Summary List API keys
// @Security AdminBearerAuth
// @Description List all API keys for a project. super_admin only.
// @Tags Admin
// @Produce json
// @Param pid path string true "Project ID"
// @Success 200 {array} auth.ApiKey
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/api-keys [get]
func (h *ApiKeysHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if claims.Role != "super_admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	projectID := r.PathValue("pid")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_project_id"})
		return
	}

	keys, err := h.Service.List(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if keys == nil {
		keys = []auth.ApiKey{}
	}

	writeJSON(w, http.StatusOK, keys)
}

// @Summary Revoke API key
// @Security AdminBearerAuth
// @Description Delete an API key. super_admin only.
// @Tags Admin
// @Param pid path string true "Project ID"
// @Param kid path string true "API key ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/api-keys/{kid} [delete]
func (h *ApiKeysHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if claims.Role != "super_admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return
	}

	projectID := r.PathValue("pid")
	keyID := r.PathValue("kid")
	if projectID == "" || keyID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	if err := h.Service.Delete(r.Context(), keyID, projectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
