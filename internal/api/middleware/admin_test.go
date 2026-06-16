package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/s-piazzano/cosmoria/internal/api/middleware"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func newAdminHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAdminAuth_ValidToken(t *testing.T) {
	cfg := testhelper.TestConfig()
	handler := middleware.AdminAuth(cfg.AdminJWTSecret)(newAdminHandler())

	token := testhelper.AdminJWT(t, cfg, "admin-uuid")

	req := httptest.NewRequest("GET", "/api/admin/projects/123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestAdminAuth_InvalidToken(t *testing.T) {
	cfg := testhelper.TestConfig()
	handler := middleware.AdminAuth(cfg.AdminJWTSecret)(newAdminHandler())

	req := httptest.NewRequest("GET", "/api/admin/projects/123", nil)
	req.Header.Set("Authorization", "Bearer invalid-admin-token")
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestAdminAuth_NoToken(t *testing.T) {
	cfg := testhelper.TestConfig()
	handler := middleware.AdminAuth(cfg.AdminJWTSecret)(newAdminHandler())

	req := httptest.NewRequest("GET", "/api/admin/projects/123", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestAdminAuth_SkipsSetup(t *testing.T) {
	cfg := testhelper.TestConfig()
	handler := middleware.AdminAuth(cfg.AdminJWTSecret)(newAdminHandler())

	req := httptest.NewRequest("POST", "/api/admin/setup", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code, "/setup should be skipped")
}

func TestAdminAuth_SkipsLogin(t *testing.T) {
	cfg := testhelper.TestConfig()
	handler := middleware.AdminAuth(cfg.AdminJWTSecret)(newAdminHandler())

	req := httptest.NewRequest("POST", "/api/admin/login", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code, "/login should be skipped")
}

func TestAdminAuth_RejectsUserJWT(t *testing.T) {
	cfg := testhelper.TestConfig()
	handler := middleware.AdminAuth(cfg.AdminJWTSecret)(newAdminHandler())

	// User JWT (signed with JWTSecret, not AdminJWTSecret)
	token := testhelper.UserJWT(t, cfg, "user-uuid", "project-uuid")

	req := httptest.NewRequest("GET", "/api/admin/projects/123", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code, "user JWT should not work for admin routes")
}
