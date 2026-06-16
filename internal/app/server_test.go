package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/app"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

type testEnv struct {
	pool    *pgxpool.Pool
	cfg     *core.Config
	server  http.Handler
	shutdown func()
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	cfg := testhelper.TestConfig()
	handler, shutdown := app.BuildHandler(cfg, pool)

	t.Cleanup(shutdown)

	return &testEnv{pool: pool, cfg: cfg, server: handler, shutdown: shutdown}
}

func TestFullFlow_SignupLoginCreateRecord(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Admin setup
	adminBody := `{"email":"admin@test.com","password":"adminpass"}`
	req := httptest.NewRequest("POST", "/api/admin/setup", strings.NewReader(adminBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var setupResult struct {
		Token   string `json:"token"`
		Admin   struct { ID string `json:"id"` } `json:"admin"`
		Project struct { ID string `json:"id"` } `json:"project"`
	}
	err := json.NewDecoder(resp.Body).Decode(&setupResult)
	require.NoError(t, err)
	require.NotEmpty(t, setupResult.Project.ID)

	// 2. User signup
	signupBody := `{"email":"user@test.com","password":"userpass","project_id":"` + setupResult.Project.ID + `"}`
	req = httptest.NewRequest("POST", "/api/auth/signup", strings.NewReader(signupBody))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var signupResult struct {
		Token string `json:"token"`
		User  struct {
			ID        string `json:"id"`
			ProjectID string `json:"project_id"`
		} `json:"user"`
	}
	err = json.NewDecoder(resp.Body).Decode(&signupResult)
	require.NoError(t, err)
	userToken := signupResult.Token
	adminToken := setupResult.Token

	// 3. Create RBAC role for the user
	roleBody := `{"name":"editor"}`
	req = httptest.NewRequest("POST", "/api/admin/projects/"+setupResult.Project.ID+"/roles", strings.NewReader(roleBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var roleResult struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&roleResult)
	require.NoError(t, err)

	// Grant full permissions
	permBody := `{"resource":"*","action":"*"}`
	req = httptest.NewRequest("POST", "/api/admin/projects/"+setupResult.Project.ID+"/roles/"+roleResult.ID+"/permissions", strings.NewReader(permBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	// Assign role to user
	assignBody := `{"role_id":"` + roleResult.ID + `"}`
	req = httptest.NewRequest("POST", "/api/admin/projects/"+setupResult.Project.ID+"/users/"+signupResult.User.ID+"/role", strings.NewReader(assignBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)

	// 4. Create tenant
	tenantBody := `{"name":"my-tenant"}`
	req = httptest.NewRequest("POST", "/api/projects/"+setupResult.Project.ID+"/tenants", strings.NewReader(tenantBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var tenantResult struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tenantResult)
	require.NoError(t, err)

	// 5. Create collection (admin)
	collBody := `{"name":"posts","schema":{"fields":[{"name":"title","type":"string","required":true}]}}`
	req = httptest.NewRequest("POST", "/api/admin/projects/"+setupResult.Project.ID+"/collections", strings.NewReader(collBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var collResult struct {
		ID string `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&collResult)
	require.NoError(t, err)

	// 6. Create record
	recordBody := `{"data":{"title":"Hello World"}}`
	req = httptest.NewRequest("POST", "/api/projects/"+setupResult.Project.ID+"/tenants/"+tenantResult.ID+"/collections/"+collResult.ID+"/records", strings.NewReader(recordBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var recordResult struct {
		ID   string `json:"id"`
		Data struct {
			Title string `json:"title"`
		} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&recordResult)
	require.NoError(t, err)
	assert.Equal(t, "Hello World", recordResult.Data.Title)

	// 7. Get record
	req = httptest.NewRequest("GET", "/api/projects/"+setupResult.Project.ID+"/tenants/"+tenantResult.ID+"/collections/"+collResult.ID+"/records/"+recordResult.ID, nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)

	// 8. Get user profile
	req = httptest.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestAPI_HealthCheck(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
}

func TestAPI_SwaggerDocs(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest("GET", "/docs/index.html", nil)
	resp := httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)

	req = httptest.NewRequest("GET", "/docs/doc.json", nil)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestAPI_UnauthenticatedAccess(t *testing.T) {
	env := setupTestEnv(t)

	protectedRoutes := []string{
		"GET /api/auth/me",
		"POST /api/projects/123/tenants",
		"POST /api/admin/projects",
	}

	for _, route := range protectedRoutes {
		parts := strings.SplitN(route, " ", 2)
		method, path := parts[0], parts[1]

		req := httptest.NewRequest(method, path, nil)
		resp := httptest.NewRecorder()
		env.server.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusUnauthorized, resp.Code, "unauthenticated %s %s should return 401", method, path)
	}
}

func TestAPI_CrossProject_NotPossible(t *testing.T) {
	env := setupTestEnv(t)

	// Sign up two users in different projects
	adminBody := `{"email":"admin1@test.com","password":"adminpass"}`
	req := httptest.NewRequest("POST", "/api/admin/setup", strings.NewReader(adminBody))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var setup struct {
		Token   string `json:"token"`
		Admin   struct { ID string `json:"id"` } `json:"admin"`
		Project struct { ID string `json:"id"` } `json:"project"`
	}
	json.NewDecoder(resp.Body).Decode(&setup)
	projectA := setup.Project.ID

	// Create another project as the same admin
	projBody := `{"name":"Project B"}`
	req = httptest.NewRequest("POST", "/api/admin/projects", strings.NewReader(projBody))
	req.Header.Set("Authorization", "Bearer "+setup.Token)
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var projectB struct { ID string `json:"id"` }
	json.NewDecoder(resp.Body).Decode(&projectB)

	// Sign up user B in project B
	signupBody := `{"email":"userb@test.com","password":"userpass","project_id":"` + projectB.ID + `"}`
	req = httptest.NewRequest("POST", "/api/auth/signup", strings.NewReader(signupBody))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	require.Equal(t, http.StatusCreated, resp.Code)

	var signupB struct { Token string `json:"token"` }
	json.NewDecoder(resp.Body).Decode(&signupB)

	// User B should not be able to access project A's tenants
	req = httptest.NewRequest("GET", "/api/projects/"+projectA+"/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+signupB.Token)
	resp = httptest.NewRecorder()
	env.server.ServeHTTP(resp, req)
	// Auth middleware matches project from JWT vs route, so this should fail
	assert.Equal(t, http.StatusForbidden, resp.Code)
}
