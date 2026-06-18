package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/tenant"
)

type AdminTenantHandler struct {
	Service *tenant.Service
}

func (h *AdminTenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_project_id"})
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

	t, err := h.Service.CreateTenant(r.Context(), projectID, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *AdminTenantHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_project_id"})
		return
	}

	tenants, err := h.Service.ListTenants(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if tenants == nil {
		tenants = []tenant.Tenant{}
	}

	writeJSON(w, http.StatusOK, tenants)
}

func (h *AdminTenantHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	tenantID := r.PathValue("tid")
	if projectID == "" || tenantID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	if err := h.Service.DeleteTenant(r.Context(), tenantID, projectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
