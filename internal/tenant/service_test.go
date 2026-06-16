package tenant_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/tenant"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupTenantTest(t *testing.T) (*pgxpool.Pool, *tenant.Service, *testhelper.TestProject, *testhelper.TestProject) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	svc := tenant.NewService(pool)

	admin := testhelper.CreateTestAdmin(t, pool)
	projectA := testhelper.CreateTestProject(t, pool, admin.ID)
	projectB := testhelper.CreateTestProject(t, pool, admin.ID)

	return pool, svc, projectA, projectB
}

func TestCreateTenant_Success(t *testing.T) {
	_, svc, project, _ := setupTenantTest(t)

	tn, err := svc.CreateTenant(context.Background(), project.ID, "My Tenant")
	require.NoError(t, err)
	assert.NotEmpty(t, tn.ID)
	assert.Equal(t, project.ID, tn.ProjectID)
	assert.Equal(t, "My Tenant", tn.Name)
}

func TestListTenants_ScopedToProject(t *testing.T) {
	_, svc, projectA, projectB := setupTenantTest(t)

	tnA, err := svc.CreateTenant(context.Background(), projectA.ID, "Tenant A")
	require.NoError(t, err)
	_, err = svc.CreateTenant(context.Background(), projectB.ID, "Tenant B")
	require.NoError(t, err)

	tenants, err := svc.ListTenants(context.Background(), projectA.ID)
	require.NoError(t, err)
	assert.Len(t, tenants, 1)
	assert.Equal(t, tnA.ID, tenants[0].ID)
}

func TestGetTenant_Success(t *testing.T) {
	_, svc, project, _ := setupTenantTest(t)

	created, err := svc.CreateTenant(context.Background(), project.ID, "Get Me")
	require.NoError(t, err)

	tn, err := svc.GetTenant(context.Background(), created.ID, project.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, tn.ID)
	assert.Equal(t, "Get Me", tn.Name)
}

func TestGetTenant_WrongProject(t *testing.T) {
	_, svc, projectA, projectB := setupTenantTest(t)

	created, err := svc.CreateTenant(context.Background(), projectA.ID, "Secret")
	require.NoError(t, err)

	_, err = svc.GetTenant(context.Background(), created.ID, projectB.ID)
	assert.Error(t, err)
}

func TestDeleteTenant_Success(t *testing.T) {
	_, svc, project, _ := setupTenantTest(t)

	created, err := svc.CreateTenant(context.Background(), project.ID, "Delete Me")
	require.NoError(t, err)

	err = svc.DeleteTenant(context.Background(), created.ID, project.ID)
	require.NoError(t, err)

	_, err = svc.GetTenant(context.Background(), created.ID, project.ID)
	assert.Error(t, err)
}

func TestDeleteTenant_WrongProject(t *testing.T) {
	_, svc, projectA, projectB := setupTenantTest(t)

	created, err := svc.CreateTenant(context.Background(), projectA.ID, "Protected")
	require.NoError(t, err)

	err = svc.DeleteTenant(context.Background(), created.ID, projectB.ID)
	assert.Error(t, err)

	_, err = svc.GetTenant(context.Background(), created.ID, projectA.ID)
	require.NoError(t, err)
}

func TestHasAccess_UserAssigned(t *testing.T) {
	pool, svc, project, _ := setupTenantTest(t)

	user := testhelper.CreateTestUser(t, pool)
	tn, err := svc.CreateTenant(context.Background(), project.ID, "Access Tenant")
	require.NoError(t, err)

	err = svc.AssignUser(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err)

	ok, err := svc.HasAccess(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestHasAccess_NotAssigned(t *testing.T) {
	pool, svc, project, _ := setupTenantTest(t)

	user := testhelper.CreateTestUser(t, pool)
	tn, err := svc.CreateTenant(context.Background(), project.ID, "No Access")
	require.NoError(t, err)

	ok, err := svc.HasAccess(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRemoveUser_RemovesAccess(t *testing.T) {
	pool, svc, project, _ := setupTenantTest(t)

	user := testhelper.CreateTestUser(t, pool)
	tn, err := svc.CreateTenant(context.Background(), project.ID, "Remove Access")
	require.NoError(t, err)

	err = svc.AssignUser(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err)

	err = svc.RemoveUser(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err)

	ok, err := svc.HasAccess(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestAssignUser_Idempotent(t *testing.T) {
	pool, svc, project, _ := setupTenantTest(t)

	user := testhelper.CreateTestUser(t, pool)
	tn, err := svc.CreateTenant(context.Background(), project.ID, "Idempotent")
	require.NoError(t, err)

	err = svc.AssignUser(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err)

	err = svc.AssignUser(context.Background(), user.ID, tn.ID, project.ID)
	require.NoError(t, err, "second assignment should be idempotent")
}
