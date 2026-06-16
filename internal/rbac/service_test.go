package rbac_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/rbac"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupRBACTest(t *testing.T) (*pgxpool.Pool, *rbac.Service, *testhelper.TestProject, *testhelper.TestUser, string) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	svc := rbac.NewService(pool)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	user := testhelper.CreateTestUser(t, pool)

	return pool, svc, project, user, user.ID
}

func createRole(t *testing.T, svc *rbac.Service, projectID, name string) string {
	t.Helper()
	role, err := svc.CreateRole(context.Background(), projectID, name)
	require.NoError(t, err)
	return role.ID
}

func TestCheckAccess_ExactMatch(t *testing.T) {
	_, svc, project, _, userID := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "editor")
	_, err := svc.SetPermission(context.Background(), roleID, "records", "create")
	require.NoError(t, err)

	_, err = svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	ok, err := svc.CheckAccess(context.Background(), userID, project.ID, "records", "create")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckAccess_ResourceWildcard(t *testing.T) {
	_, svc, project, _, userID := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "admin")
	_, err := svc.SetPermission(context.Background(), roleID, "*", "read")
	require.NoError(t, err)

	_, err = svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	ok, err := svc.CheckAccess(context.Background(), userID, project.ID, "records", "read")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = svc.CheckAccess(context.Background(), userID, project.ID, "files", "read")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckAccess_ActionWildcard(t *testing.T) {
	_, svc, project, _, userID := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "super-editor")
	_, err := svc.SetPermission(context.Background(), roleID, "records", "*")
	require.NoError(t, err)

	_, err = svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	ok, err := svc.CheckAccess(context.Background(), userID, project.ID, "records", "create")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = svc.CheckAccess(context.Background(), userID, project.ID, "records", "delete")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckAccess_BothWildcards(t *testing.T) {
	_, svc, project, _, userID := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "super-admin")
	_, err := svc.SetPermission(context.Background(), roleID, "*", "*")
	require.NoError(t, err)

	_, err = svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	ok, err := svc.CheckAccess(context.Background(), userID, project.ID, "tenants", "create")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = svc.CheckAccess(context.Background(), userID, project.ID, "files", "delete")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCheckAccess_Denied(t *testing.T) {
	_, svc, project, _, userID := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "reader")
	_, err := svc.SetPermission(context.Background(), roleID, "records", "read")
	require.NoError(t, err)

	_, err = svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	ok, err := svc.CheckAccess(context.Background(), userID, project.ID, "records", "delete")
	require.NoError(t, err)
	assert.False(t, ok)

	ok, err = svc.CheckAccess(context.Background(), userID, project.ID, "files", "read")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckAccess_NoRole(t *testing.T) {
	_, svc, project, _, userID := setupRBACTest(t)

	ok, err := svc.CheckAccess(context.Background(), userID, project.ID, "records", "read")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCheckAccess_AnotherUserDenied(t *testing.T) {
	pool, svc, project, _, userID := setupRBACTest(t)

	otherUser := testhelper.CreateTestUser(t, pool)

	roleID := createRole(t, svc, project.ID, "editor")
	_, err := svc.SetPermission(context.Background(), roleID, "records", "create")
	require.NoError(t, err)

	_, err = svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)

	ok, err := svc.CheckAccess(context.Background(), otherUser.ID, project.ID, "records", "create")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRoleCRUD(t *testing.T) {
	_, svc, project, _, _ := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "manager")

	roles, err := svc.ListRoles(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, "manager", roles[0].Name)

	err = svc.DeleteRole(context.Background(), roleID, project.ID)
	require.NoError(t, err)

	roles, err = svc.ListRoles(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 0)
}

func TestPermissionManagement(t *testing.T) {
	_, svc, project, _, _ := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "tester")

	perm, err := svc.SetPermission(context.Background(), roleID, "files", "upload")
	require.NoError(t, err)
	assert.Equal(t, "files", perm.Resource)
	assert.Equal(t, "upload", perm.Action)

	perms, err := svc.ListPermissions(context.Background(), roleID)
	require.NoError(t, err)
	assert.Len(t, perms, 1)

	err = svc.RemovePermission(context.Background(), roleID, "files", "upload")
	require.NoError(t, err)

	perms, err = svc.ListPermissions(context.Background(), roleID)
	require.NoError(t, err)
	assert.Len(t, perms, 0)
}

func TestUserRoleAssignment(t *testing.T) {
	_, svc, project, _, userID := setupRBACTest(t)

	roleID := createRole(t, svc, project.ID, "member")

	ur, err := svc.AssignUserRole(context.Background(), userID, project.ID, roleID)
	require.NoError(t, err)
	assert.Equal(t, userID, ur.UserID)
	assert.Equal(t, project.ID, ur.ProjectID)

	got, err := svc.GetUserRole(context.Background(), userID, project.ID)
	require.NoError(t, err)
	assert.Equal(t, roleID, got.RoleID)

	err = svc.RemoveUserRole(context.Background(), userID, project.ID)
	require.NoError(t, err)

	_, err = svc.GetUserRole(context.Background(), userID, project.ID)
	assert.Error(t, err)
}
