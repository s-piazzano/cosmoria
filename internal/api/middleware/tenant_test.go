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
	"github.com/s-piazzano/cosmoria/internal/tenant"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupTenantMiddlewareTest(t *testing.T) (*pgxpool.Pool, *tenant.Service, *testhelper.TestProject, string) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	svc := tenant.NewService(pool)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	user := testhelper.CreateTestUser(t, pool)

	return pool, svc, project, user.ID
}

func TestTenantMiddleware_WithValidAccess(t *testing.T) {
	_, svc, project, userID := setupTenantMiddlewareTest(t)

	tn, err := svc.CreateTenant(context.Background(), project.ID, "my-tenant")
	require.NoError(t, err)

	err = svc.AssignUser(context.Background(), userID, tn.ID, project.ID)
	require.NoError(t, err)

	handler := middleware.Tenant(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/projects/"+project.ID+"/tenants/"+tn.ID+"/records", nil)
	req.Header.Set("X-Tenant-ID", tn.ID)
	req = injectUserClaims(req, userID, project.ID)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestTenantMiddleware_WithoutAccess(t *testing.T) {
	_, svc, project, userID := setupTenantMiddlewareTest(t)

	tn, err := svc.CreateTenant(context.Background(), project.ID, "restricted")
	require.NoError(t, err)

	handler := middleware.Tenant(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/projects/"+project.ID+"/tenants/"+tn.ID+"/records", nil)
	req.Header.Set("X-Tenant-ID", tn.ID)
	req = injectUserClaims(req, userID, project.ID)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestTenantMiddleware_NoHeader(t *testing.T) {
	_, svc, project, userID := setupTenantMiddlewareTest(t)

	handler := middleware.Tenant(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/projects/"+project.ID+"/tenants", nil)
	req = injectUserClaims(req, userID, project.ID)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code, "no X-Tenant-ID should pass through")
}

func TestTenantMiddleware_SkipsPublicRoutes(t *testing.T) {
	_, svc, _, _ := setupTenantMiddlewareTest(t)

	handler := middleware.Tenant(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
