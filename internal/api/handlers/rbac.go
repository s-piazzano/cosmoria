package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/rbac"
)

type RolesHandler struct {
	Service *rbac.Service
}

type createRoleRequest struct {
	Name string `json:"name"`
}

type permissionRequest struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type userRoleRequest struct {
	RoleID string `json:"role_id"`
}

func (h *RolesHandler) checkSuperAdmin(w http.ResponseWriter, r *http.Request) bool {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return false
	}
	if claims.Role != "super_admin" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
		return false
	}
	return true
}

// @Summary Create a role
// @Security AdminBearerAuth
// @Description Create a new RBAC role for a project. super_admin only.
// @Tags RBAC
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param body body createRoleRequest true "Role name"
// @Success 201 {object} rbac.Role
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/roles [post]
func (h *RolesHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	projectID := r.PathValue("pid")

	var req createRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name_required"})
		return
	}

	role, err := h.Service.CreateRole(r.Context(), projectID, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
		return
	}

	writeJSON(w, http.StatusCreated, role)
}

// @Summary List roles
// @Security AdminBearerAuth
// @Description List all RBAC roles for a project. super_admin only.
// @Tags RBAC
// @Produce json
// @Param pid path string true "Project ID"
// @Success 200 {array} rbac.RoleWithPermissions
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/roles [get]
func (h *RolesHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	projectID := r.PathValue("pid")

	roles, err := h.Service.ListRoles(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if roles == nil {
		roles = []rbac.RoleWithPermissions{}
	}

	writeJSON(w, http.StatusOK, roles)
}

// @Summary Delete a role
// @Security AdminBearerAuth
// @Description Delete a role and its permissions. super_admin only.
// @Tags RBAC
// @Param pid path string true "Project ID"
// @Param rid path string true "Role ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/roles/{rid} [delete]
func (h *RolesHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	projectID := r.PathValue("pid")
	roleID := r.PathValue("rid")

	if err := h.Service.DeleteRole(r.Context(), roleID, projectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Set a permission on a role
// @Security AdminBearerAuth
// @Description Add a (resource, action) permission to a role. super_admin only.
// @Tags RBAC
// @Accept json
// @Produce json
// @Param rid path string true "Role ID"
// @Param body body permissionRequest true "Resource and action"
// @Success 201 {object} rbac.Permission
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/roles/{rid}/permissions [post]
func (h *RolesHandler) SetPermission(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	roleID := r.PathValue("rid")

	var req permissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Resource == "" || req.Action == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resource_and_action_required"})
		return
	}

	perm, err := h.Service.SetPermission(r.Context(), roleID, req.Resource, req.Action)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "set_permission_failed"})
		return
	}

	writeJSON(w, http.StatusCreated, perm)
}

// @Summary Remove a permission from a role
// @Security AdminBearerAuth
// @Description Remove a (resource, action) permission from a role. super_admin only.
// @Tags RBAC
// @Accept json
// @Param rid path string true "Role ID"
// @Param body body permissionRequest true "Resource and action"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/roles/{rid}/permissions [delete]
func (h *RolesHandler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	roleID := r.PathValue("rid")

	var req permissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Resource == "" || req.Action == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "resource_and_action_required"})
		return
	}

	if err := h.Service.RemovePermission(r.Context(), roleID, req.Resource, req.Action); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "remove_permission_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary List permissions for a role
// @Security AdminBearerAuth
// @Description List all permissions assigned to a role. super_admin only.
// @Tags RBAC
// @Produce json
// @Param rid path string true "Role ID"
// @Success 200 {array} rbac.Permission
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/roles/{rid}/permissions [get]
func (h *RolesHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	roleID := r.PathValue("rid")

	perms, err := h.Service.ListPermissions(r.Context(), roleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if perms == nil {
		perms = []rbac.Permission{}
	}

	writeJSON(w, http.StatusOK, perms)
}

// @Summary Assign a role to a user
// @Security AdminBearerAuth
// @Description Assign an RBAC role to a SaaS user for a project. super_admin only.
// @Tags RBAC
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param uid path string true "User ID"
// @Param body body userRoleRequest true "Role ID"
// @Success 200 {object} rbac.UserProjectRole
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/users/{uid}/role [post]
func (h *RolesHandler) AssignUserRole(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	projectID := r.PathValue("pid")
	userID := r.PathValue("uid")

	var req userRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.RoleID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role_id_required"})
		return
	}

	upr, err := h.Service.AssignUserRole(r.Context(), userID, projectID, req.RoleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "assign_failed"})
		return
	}

	writeJSON(w, http.StatusOK, upr)
}

// @Summary Get user role
// @Security AdminBearerAuth
// @Description Get the RBAC role assigned to a SaaS user for a project. super_admin only.
// @Tags RBAC
// @Produce json
// @Param pid path string true "Project ID"
// @Param uid path string true "User ID"
// @Success 200 {object} rbac.UserProjectRole
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/admin/projects/{pid}/users/{uid}/role [get]
func (h *RolesHandler) GetUserRole(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	projectID := r.PathValue("pid")
	userID := r.PathValue("uid")

	upr, err := h.Service.GetUserRole(r.Context(), userID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "user_role_not_found"})
		return
	}

	writeJSON(w, http.StatusOK, upr)
}

// @Summary Remove user role
// @Security AdminBearerAuth
// @Description Remove the RBAC role assignment from a SaaS user. super_admin only.
// @Tags RBAC
// @Param pid path string true "Project ID"
// @Param uid path string true "User ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/admin/projects/{pid}/users/{uid}/role [delete]
func (h *RolesHandler) RemoveUserRole(w http.ResponseWriter, r *http.Request) {
	if !h.checkSuperAdmin(w, r) {
		return
	}
	projectID := r.PathValue("pid")
	userID := r.PathValue("uid")

	if err := h.Service.RemoveUserRole(r.Context(), userID, projectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "remove_failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
