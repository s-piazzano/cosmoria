package auth_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupAuthTest(t *testing.T) (*pgxpool.Pool, *auth.Service, *testhelper.TestAdmin, *testhelper.TestProject) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	cfg := testhelper.TestConfig()

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)

	svc := auth.NewService(pool, cfg)
	return pool, svc, admin, project
}

func TestSignup_Success(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	result, err := svc.Signup(context.Background(), auth.SignupInput{
		Email:     "new@test.com",
		Password:  "securepass123",
		ProjectID: project.ID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "new@test.com", result.User.Email)
	assert.NotEmpty(t, result.User.ID)
	assert.Equal(t, project.ID, result.User.ProjectID)
}

func TestSignup_DuplicateEmail(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	_, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "dup@test.com", Password: "pass1", ProjectID: project.ID,
	})
	require.NoError(t, err)

	_, err = svc.Signup(context.Background(), auth.SignupInput{
		Email: "dup@test.com", Password: "pass2", ProjectID: project.ID,
	})
	assert.Error(t, err)
}

func TestSignup_InvalidProject(t *testing.T) {
	_, svc, _, _ := setupAuthTest(t)

	_, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "no@project.com", Password: "pass", ProjectID: "00000000-0000-0000-0000-000000000000",
	})
	assert.Error(t, err)
}

func TestLogin_Success(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	email := "login@test.com"
	password := "mypassword"

	_, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: email, Password: password, ProjectID: project.ID,
	})
	require.NoError(t, err)

	result, err := svc.Login(context.Background(), auth.LoginInput{
		Email: email, Password: password, ProjectID: project.ID,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, email, result.User.Email)
}

func TestLogin_WrongPassword(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	_, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "wrong@test.com", Password: "correctpass", ProjectID: project.ID,
	})
	require.NoError(t, err)

	_, err = svc.Login(context.Background(), auth.LoginInput{
		Email: "wrong@test.com", Password: "wrongpass", ProjectID: project.ID,
	})
	assert.Error(t, err)
}

func TestLogin_NonExistentUser(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	_, err := svc.Login(context.Background(), auth.LoginInput{
		Email: "nonexistent@test.com", Password: "pass", ProjectID: project.ID,
	})
	assert.Error(t, err)
}

func TestGetByID_Success(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	signup, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "getbyid@test.com", Password: "pass", ProjectID: project.ID,
	})
	require.NoError(t, err)

	user, err := svc.GetByID(context.Background(), signup.User.ID)
	require.NoError(t, err)
	assert.Equal(t, signup.User.ID, user.ID)
	assert.Equal(t, "getbyid@test.com", user.Email)
}

func TestGetByID_NotFound(t *testing.T) {
	_, svc, _, _ := setupAuthTest(t)

	_, err := svc.GetByID(context.Background(), "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestUpdateEmail_Success(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	signup, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "old@test.com", Password: "pass", ProjectID: project.ID,
	})
	require.NoError(t, err)

	user, err := svc.UpdateEmail(context.Background(), auth.UpdateEmailInput{
		UserID: signup.User.ID,
		Email:  "new@test.com",
	})
	require.NoError(t, err)
	assert.Equal(t, "new@test.com", user.Email)
}

func TestUpdatePassword_Success(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	signup, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "pwdup@test.com", Password: "oldpass", ProjectID: project.ID,
	})
	require.NoError(t, err)

	err = svc.UpdatePassword(context.Background(), auth.UpdatePasswordInput{
		UserID:          signup.User.ID,
		CurrentPassword: "oldpass",
		NewPassword:     "newpass",
	})
	require.NoError(t, err)

	result, err := svc.Login(context.Background(), auth.LoginInput{
		Email: "pwdup@test.com", Password: "newpass", ProjectID: project.ID,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, result.Token)
}

func TestUpdatePassword_WrongCurrent(t *testing.T) {
	_, svc, _, project := setupAuthTest(t)

	signup, err := svc.Signup(context.Background(), auth.SignupInput{
		Email: "wrongpw@test.com", Password: "correct", ProjectID: project.ID,
	})
	require.NoError(t, err)

	err = svc.UpdatePassword(context.Background(), auth.UpdatePasswordInput{
		UserID:          signup.User.ID,
		CurrentPassword: "wrong",
		NewPassword:     "new",
	})
	assert.Error(t, err)
}
