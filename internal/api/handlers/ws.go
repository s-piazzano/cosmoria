package handlers

import (
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/realtime"
	"github.com/s-piazzano/cosmoria/internal/tenant"
)

func newUpgrader(allowedOrigins []string) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			if len(allowedOrigins) == 0 {
				return origin == "https://"+r.Host || origin == "http://"+r.Host
			}
			for _, o := range allowedOrigins {
				if strings.EqualFold(origin, o) {
					return true
				}
			}
			return false
		},
	}
}

type WSHandler struct {
	Hub            *realtime.Hub
	JWTSecret      string
	TenantService  *tenant.Service
	AllowedOrigins []string
}

// @Summary WebSocket realtime
// @Description Upgrade to WebSocket for realtime events. Authenticate via ?token= query param. Requires Bearer JWT.
// @Tags Realtime
// @Param pid path string true "Project ID"
// @Param token query string true "JWT token"
// @Success 101 "Switching Protocols"
// @Failure 401 {object} map[string]string
// @Router /api/projects/{pid}/ws [get]
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "token_required"})
		return
	}

	claims, err := auth.ValidateToken(tokenStr, h.JWTSecret)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_token"})
		return
	}

	projectID := r.PathValue("pid")
	if projectID == "" || projectID != claims.ProjectID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "project_mismatch"})
		return
	}

	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID != "" {
		ok, err := h.TenantService.HasAccess(r.Context(), claims.UserID, tenantID, projectID)
		if err != nil || !ok {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "tenant_access_denied"})
			return
		}
	}

	upgrader := newUpgrader(h.AllowedOrigins)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := realtime.NewClient(h.Hub, conn, claims.UserID, projectID, tenantID)
	client.Start()
}
