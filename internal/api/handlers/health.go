package handlers

import (
	"net/http"
)

// @Summary Health check
// @Description Returns the health status of the server.
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "status: ok"
// @Router /health [get]
func Health(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
