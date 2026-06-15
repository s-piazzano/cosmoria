package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
)

type AdminHandler struct {
	Service *adminauth.Service
}

type setupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createProjectRequest struct {
	Name string `json:"name"`
}

type assignRoleRequest struct {
	AdminUserID string `json:"admin_user_id"`
	Role        string `json:"role"`
}

// @Summary Bootstrap the platform
// @Description Create the first super_admin and default project. Only works once.
// @Tags Admin
// @Accept json
// @Produce json
// @Param body body setupRequest true "Admin credentials"
// @Success 201 {object} map[string]any "token, admin, project"
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/admin/setup [post]
func (h *AdminHandler) Setup(w http.ResponseWriter, r *http.Request) {
	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
		return
	}

	result, project, err := h.Service.Setup(r.Context(), req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"token":   result.Token,
		"admin":   result.Admin,
		"project": project,
	})
}

// @Summary Admin login
// @Description Authenticate as a platform admin and receive a JWT.
// @Tags Admin
// @Accept json
// @Produce json
// @Param body body loginRequest true "Admin credentials"
// @Success 200 {object} adminauth.AuthResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/admin/login [post]
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
		return
	}

	result, err := h.Service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// @Summary Create a project
// @Security AdminBearerAuth
// @Description Create a new project as a platform admin.
// @Tags Admin
// @Accept json
// @Produce json
// @Param body body createProjectRequest true "Project name"
// @Success 201 {object} adminauth.Project
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/projects [post]
func (h *AdminHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	var req createProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name_required"})
		return
	}

	project, err := h.Service.CreateProject(r.Context(), claims.AdminUserID, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

// @Summary List accessible projects
// @Security AdminBearerAuth
// @Description List all projects accessible by the authenticated admin.
// @Tags Admin
// @Produce json
// @Success 200 {array} adminauth.ProjectWithRole
// @Failure 500 {object} map[string]string
// @Router /api/admin/projects [get]
func (h *AdminHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projects, err := h.Service.ListAccessibleProjects(r.Context(), claims.AdminUserID, claims.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if projects == nil {
		projects = []adminauth.ProjectWithRole{}
	}

	writeJSON(w, http.StatusOK, projects)
}

// @Summary Assign admin role to a project
// @Security AdminBearerAuth
// @Description Assign an admin user to a project with a specific role. super_admin only.
// @Tags Admin
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param body body assignRoleRequest true "Admin user ID and role"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/admin-roles [post]
func (h *AdminHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
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

	var req assignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.AdminUserID == "" || req.Role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
		return
	}

	if err := h.Service.AssignRole(r.Context(), projectID, req.AdminUserID, req.Role); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "assign_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Remove admin role from a project
// @Security AdminBearerAuth
// @Description Remove an admin user's access to a project. super_admin only.
// @Tags Admin
// @Param pid path string true "Project ID"
// @Param aid path string true "Admin user ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/admin-roles/{aid} [delete]
func (h *AdminHandler) RemoveRole(w http.ResponseWriter, r *http.Request) {
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
	adminUserID := r.PathValue("aid")
	if projectID == "" || adminUserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	if err := h.Service.RemoveRole(r.Context(), projectID, adminUserID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "remove_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary List admin roles for a project
// @Security AdminBearerAuth
// @Description List all admin roles assigned to a project. super_admin only.
// @Tags Admin
// @Produce json
// @Param pid path string true "Project ID"
// @Success 200 {array} adminauth.AdminProjectRole
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/admin-roles [get]
func (h *AdminHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
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

	roles, err := h.Service.ListRoles(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if roles == nil {
		roles = []adminauth.AdminProjectRole{}
	}

	writeJSON(w, http.StatusOK, roles)
}
