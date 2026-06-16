package adminauth_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupAdminTest(t *testing.T) (*pgxpool.Pool, *adminauth.Service, *core.Config) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	cfg := testhelper.TestConfig()
	svc := adminauth.NewService(pool, cfg)

	return pool, svc, cfg
}

func TestAdmin_Setup_FirstTime(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	result, project, err := svc.Setup(context.Background(), "setup@test.com", "password", "Default Project")
	require.NoError(t, err)

	require.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "setup@test.com", result.Admin.Email)
	assert.Equal(t, "super_admin", result.Admin.Role)

	require.NotNil(t, project)
	assert.Equal(t, "Default Project", project.Name)
	assert.NotEmpty(t, project.ID)
}

func TestAdmin_Setup_SecondTimeFails(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	_, _, err := svc.Setup(context.Background(), "admin1@test.com", "pass", "Proj 1")
	require.NoError(t, err)

	_, _, err = svc.Setup(context.Background(), "admin2@test.com", "pass", "Proj 2")
	assert.Error(t, err, "setup should fail when admin_users already exist")
}

func TestAdmin_Login_Success(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	email := "login-admin@test.com"
	password := "adminpass"

	_, _, err := svc.Setup(context.Background(), email, password, "Login Project")
	require.NoError(t, err)

	result, err := svc.Login(context.Background(), email, password)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, email, result.Admin.Email)
}

func TestAdmin_Login_WrongPassword(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	_, _, err := svc.Setup(context.Background(), "wrong@test.com", "correct", "Proj")
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), "wrong@test.com", "wrong")
	assert.Error(t, err)
}

func TestAdmin_CreateProject(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	result, _, err := svc.Setup(context.Background(), "create-proj@test.com", "pass", "Initial")
	require.NoError(t, err)

	project, err := svc.CreateProject(context.Background(), result.Admin.ID, "New Project")
	require.NoError(t, err)
	assert.Equal(t, "New Project", project.Name)
	assert.NotEmpty(t, project.ID)
}

func TestAdmin_GetProject(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	result, initialProject, err := svc.Setup(context.Background(), "get-proj@test.com", "pass", "My Project")
	require.NoError(t, err)

	project, err := svc.GetProject(context.Background(), initialProject.ID)
	require.NoError(t, err)
	assert.Equal(t, "My Project", project.Name)

	newProj, err := svc.CreateProject(context.Background(), result.Admin.ID, "Another")
	require.NoError(t, err)

	p, err := svc.GetProject(context.Background(), newProj.ID)
	require.NoError(t, err)
	assert.Equal(t, "Another", p.Name)
}

func TestAdmin_UpdateProject(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	_, project, err := svc.Setup(context.Background(), "update@test.com", "pass", "Old Name")
	require.NoError(t, err)

	updated, err := svc.UpdateProject(context.Background(), project.ID, adminauth.UpdateProjectInput{
		Name: "New Name",
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
}

func TestAdmin_DeleteProject_Cascade(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	result, project, err := svc.Setup(context.Background(), "delete@test.com", "pass", "To Delete")
	require.NoError(t, err)

	// Add dependent data
	_, err = svc.CreateProject(context.Background(), result.Admin.ID, "Stays")
	require.NoError(t, err)

	err = svc.DeleteProject(context.Background(), project.ID)
	require.NoError(t, err)

	// Verify project is gone
	_, err = svc.GetProject(context.Background(), project.ID)
	assert.Error(t, err, "deleted project should not exist")

	// Other projects remain
	projects, err := svc.ListAccessibleProjects(context.Background(), result.Admin.ID, "super_admin")
	require.NoError(t, err)
	assert.Len(t, projects, 1, "only remaining project should be listed")
}

func TestAdmin_ListAccessibleProjects_SuperAdmin(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	result, _, err := svc.Setup(context.Background(), "super@test.com", "pass", "First")
	require.NoError(t, err)

	_, err = svc.CreateProject(context.Background(), result.Admin.ID, "Second")
	require.NoError(t, err)

	projects, err := svc.ListAccessibleProjects(context.Background(), result.Admin.ID, "super_admin")
	require.NoError(t, err)
	assert.Len(t, projects, 2)
}

func TestAdmin_AssignAndRemoveRole(t *testing.T) {
	_, svc, _ := setupAdminTest(t)

	result, project, err := svc.Setup(context.Background(), "assign@test.com", "pass", "Assign Project")
	require.NoError(t, err)

	err = svc.AssignRole(context.Background(), project.ID, result.Admin.ID, "admin")
	require.NoError(t, err)

	roles, err := svc.ListRoles(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 1)

	err = svc.RemoveRole(context.Background(), project.ID, result.Admin.ID)
	require.NoError(t, err)

	roles, err = svc.ListRoles(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Len(t, roles, 0)
}


