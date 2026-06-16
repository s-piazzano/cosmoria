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
	"github.com/s-piazzano/cosmoria/internal/rbac"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupRBACMiddlewareTest(t *testing.T) (*pgxpool.Pool, *rbac.Service, *testhelper.TestProject, string, string) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	svc := rbac.NewService(pool)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	user := testhelper.CreateTestUser(t, pool)

	return pool, svc, project, user.ID, admin.ID
}

func createRoleWithPermission(t *testing.T, svc *rbac.Service, projectID, resource, action string) string {
	t.Helper()
	role, err := svc.CreateRole(context.Background(), projectID, "tester")
	require.NoError(t, err)
	_, err = svc.SetPermission(context.Background(), role.ID, resource, action)
	require.NoError(t, err)
	return role.ID
}

func TestRequirePermission_Granted(t *testing.T) {
	_, svc, project, userID, _ := setupRBACMiddlewareTest(t)

	roleID := createRoleWithPermission(t, svc, project.ID, "records", "create")
	_, err := svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	handler := middleware.RequirePermission(svc, "records", "create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/projects/"+project.ID+"/tenants/xxx/collections/yyy/records", nil)
	req = injectUserClaims(req, userID, project.ID)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRequirePermission_Denied(t *testing.T) {
	_, svc, project, userID, _ := setupRBACMiddlewareTest(t)

	roleID := createRoleWithPermission(t, svc, project.ID, "records", "read")
	_, err := svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	handler := middleware.RequirePermission(svc, "records", "delete")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("DELETE", "/api/projects/"+project.ID+"/records/xxx", nil)
	req = injectUserClaims(req, userID, project.ID)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusForbidden, resp.Code)
}

func TestRequirePermission_NoAuthClaims(t *testing.T) {
	_, svc, _, _, _ := setupRBACMiddlewareTest(t)

	handler := middleware.RequirePermission(svc, "records", "create")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/api/projects/xxx/records", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestRequirePermission_WildcardResource(t *testing.T) {
	_, svc, project, userID, _ := setupRBACMiddlewareTest(t)

	roleID := createRoleWithPermission(t, svc, project.ID, "*", "read")
	_, err := svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	handler := middleware.RequirePermission(svc, "files", "read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/projects/"+project.ID+"/files/xxx", nil)
	req = injectUserClaims(req, userID, project.ID)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRequirePermission_WildcardAction(t *testing.T) {
	_, svc, project, userID, _ := setupRBACMiddlewareTest(t)

	roleID := createRoleWithPermission(t, svc, project.ID, "records", "*")
	_, err := svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	handler := middleware.RequirePermission(svc, "records", "update")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("PUT", "/api/projects/"+project.ID+"/records/xxx", nil)
	req = injectUserClaims(req, userID, project.ID)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
