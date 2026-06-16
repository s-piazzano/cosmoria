package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/api/middleware"
	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func newTestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := auth.GetAuth(r.Context())
		if claims != nil {
			w.Header().Set("X-User-Id", claims.UserID)
		}
		w.WriteHeader(http.StatusOK)
	})
}

func setupAuthMiddlewareTest(t *testing.T) (*pgxpool.Pool, *auth.ApiKeyService, *core.Config) {
	t.Helper()
	pool := testhelper.NewTestDB(t)
	cfg := testhelper.TestConfig()
	apiKeySvc := auth.NewApiKeyService(pool)
	return pool, apiKeySvc, cfg
}

func TestAuthMiddleware_ValidJWT(t *testing.T) {
	pool, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	// Create a valid JWT via helper
	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	user := testhelper.CreateTestUser(t, pool)
	token := testhelper.UserJWT(t, cfg, user.ID, project.ID)

	req := httptest.NewRequest("GET", "/api/projects/some-id/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, user.ID, resp.Header().Get("X-User-Id"))
}

func TestAuthMiddleware_InvalidJWT(t *testing.T) {
	_, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	req := httptest.NewRequest("GET", "/api/projects/some-id/tenants", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestAuthMiddleware_NoAuth(t *testing.T) {
	_, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	req := httptest.NewRequest("GET", "/api/projects/some-id/tenants", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestAuthMiddleware_PublicRoutes(t *testing.T) {
	_, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	publicPaths := []string{
		"/health",
		"/openapi.json",
		"/docs/",
		"/docs/doc.json",
		"/api/auth/signup",
		"/api/auth/login",
	}

	for _, path := range publicPaths {
		req := httptest.NewRequest("GET", path, nil)
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code, "public route should pass: %s", path)
	}
}

func TestAuthMiddleware_AdminRoutes(t *testing.T) {
	_, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	req := httptest.NewRequest("GET", "/api/admin/projects", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code, "admin routes should be skipped by user auth")
}

func TestAuthMiddleware_WebSocketRoute(t *testing.T) {
	_, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	req := httptest.NewRequest("GET", "/api/projects/test/ws", nil)
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code, "WS routes should be skipped by middleware")
}

func TestAuthMiddleware_ValidApiKey(t *testing.T) {
	pool, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	user := testhelper.CreateTestUser(t, pool)

	result, err := apiKeySvc.Create(context.Background(), project.ID, user.ID, "test-key", nil)
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/projects/"+project.ID+"/tenants", nil)
	req.Header.Set("X-Api-Key", result.PlainKey)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, user.ID, resp.Header().Get("X-User-Id"))
}

func TestAuthMiddleware_InvalidApiKey(t *testing.T) {
	_, apiKeySvc, cfg := setupAuthMiddlewareTest(t)

	handler := middleware.Auth(cfg.JWTSecret, apiKeySvc)(newTestHandler())

	req := httptest.NewRequest("GET", "/api/projects/some/tenants", nil)
	req.Header.Set("X-Api-Key", "ck_invalidkey0000000000000000000000000000000000000000000000000")
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}
