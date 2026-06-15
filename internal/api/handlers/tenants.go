package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/tenant"
)

type TenantHandler struct {
	Service *tenant.Service
}

type createTenantRequest struct {
	Name string `json:"name"`
}

type assignUserRequest struct {
	UserID string `json:"user_id"`
}

// @Summary Create a tenant
// @Security BearerAuth
// @Description Create a new tenant within a project.
// @Tags Tenants
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param body body createTenantRequest true "Tenant name"
// @Success 201 {object} tenant.Tenant
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects/{pid}/tenants [post]
func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req createTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name_required"})
		return
	}

	t, err := h.Service.CreateTenant(r.Context(), claims.ProjectID, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

// @Summary List tenants
// @Security BearerAuth
// @Description List all tenants within a project.
// @Tags Tenants
// @Produce json
// @Param pid path string true "Project ID"
// @Success 200 {array} tenant.Tenant
// @Failure 500 {object} map[string]string
// @Router /api/projects/{pid}/tenants [get]
func (h *TenantHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenants, err := h.Service.ListTenants(r.Context(), claims.ProjectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if tenants == nil {
		tenants = []tenant.Tenant{}
	}

	writeJSON(w, http.StatusOK, tenants)
}

// @Summary Get a tenant
// @Security BearerAuth
// @Description Get tenant details by ID.
// @Tags Tenants
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Success 200 {object} tenant.Tenant
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid} [get]
func (h *TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_tenant_id"})
		return
	}

	t, err := h.Service.GetTenant(r.Context(), tenantID, claims.ProjectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "tenant_not_found"})
		return
	}

	writeJSON(w, http.StatusOK, t)
}

// @Summary Delete a tenant
// @Security BearerAuth
// @Description Delete a tenant and its data.
// @Tags Tenants
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid} [delete]
func (h *TenantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_tenant_id"})
		return
	}

	if err := h.Service.DeleteTenant(r.Context(), tenantID, claims.ProjectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Assign user to tenant
// @Security BearerAuth
// @Description Assign an existing user to a tenant.
// @Tags Tenants
// @Accept json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param body body assignUserRequest true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid}/users [post]
func (h *TenantHandler) AssignUser(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	if tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_tenant_id"})
		return
	}

	var req assignUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id_required"})
		return
	}

	if err := h.Service.AssignUser(r.Context(), req.UserID, tenantID, claims.ProjectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "assign_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Remove user from tenant
// @Security BearerAuth
// @Description Remove a user's access to a tenant.
// @Tags Tenants
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param uid path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid}/users/{uid} [delete]
func (h *TenantHandler) RemoveUser(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	userID := r.PathValue("uid")
	if tenantID == "" || userID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	if err := h.Service.RemoveUser(r.Context(), userID, tenantID, claims.ProjectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "remove_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
