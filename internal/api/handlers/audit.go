package handlers

import (
	"net/http"
	"strconv"

	"github.com/s-piazzano/cosmoria/internal/audit"
)

type AuditHandler struct {
	Service *audit.Service
}

// List returns paginated audit logs for a project.
// @Summary List audit logs
// @Description List all audit log entries for a project with cursor-based pagination. Admin only.
// @Tags Admin
// @Produce json
// @Param pid path string true "Project ID"
// @Param cursor query string false "Pagination cursor (created_at timestamp)"
// @Param limit query int false "Max results (1-100)"
// @Success 200 {object} map[string]any
// @Failure 500 {object} map[string]string
// @Security AdminBearerAuth
// @Router /api/admin/projects/{pid}/audit-logs [get]
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("pid")
	cursor := r.URL.Query().Get("cursor")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	entries, nextCursor, err := h.Service.List(r.Context(), projectID, cursor, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"audit_logs":  entries,
		"next_cursor": nextCursor,
	})
}
