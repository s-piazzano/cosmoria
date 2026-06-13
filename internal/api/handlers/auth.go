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
