package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/auth"
)

type AuthHandler struct {
	Service *auth.Service
}

type authRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	ProjectID string `json:"project_id"`
}

// @Summary Register a new SaaS user
// @Description Create a new end-user account scoped to a project.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body authRequest true "Signup credentials"
// @Success 201 {object} auth.AuthResult
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/auth/signup [post]
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	if req.Email == "" || req.Password == "" || req.ProjectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
		return
	}

	result, err := h.Service.Signup(r.Context(), auth.SignupInput{
		Email:     req.Email,
		Password:  req.Password,
		ProjectID: req.ProjectID,
	})
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "email_already_exists"})
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

// @Summary Login as a SaaS user
// @Description Authenticate and receive a JWT token for API access.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body authRequest true "Login credentials"
// @Success 200 {object} auth.AuthResult
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	if req.Email == "" || req.Password == "" || req.ProjectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
		return
	}

	result, err := h.Service.Login(r.Context(), auth.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		ProjectID: req.ProjectID,
	})
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_credentials"})
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
