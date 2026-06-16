package auth_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupApiKeyTest(t *testing.T) (*pgxpool.Pool, *auth.ApiKeyService, *testhelper.TestProject, *testhelper.TestUser) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	svc := auth.NewApiKeyService(pool)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	user := testhelper.CreateTestUser(t, pool)

	return pool, svc, project, user
}

func TestApiKey_Create_Format(t *testing.T) {
	_, svc, project, user := setupApiKeyTest(t)

	result, err := svc.Create(context.Background(), project.ID, user.ID, "my-key", nil)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(result.PlainKey, "ck_"), "key should start with ck_")
	assert.Len(t, result.PlainKey, 3+64, "key should be ck_ + 64 hex chars")
	assert.Equal(t, "my-key", result.ApiKey.Name)
	assert.Equal(t, project.ID, result.ApiKey.ProjectID)
	assert.Equal(t, user.ID, result.ApiKey.UserID)

	// Verify the hash is SHA-256 of the plain key
	hash := sha256.Sum256([]byte(result.PlainKey))
	expectedHash := hex.EncodeToString(hash[:])
	assert.Equal(t, expectedHash[:8], result.ApiKey.KeyPrefix, "prefix should be first 8 chars of hash")
}

func TestApiKey_Validate_Success(t *testing.T) {
	_, svc, project, user := setupApiKeyTest(t)

	result, err := svc.Create(context.Background(), project.ID, user.ID, "validate-key", nil)
	require.NoError(t, err)

	claims, err := svc.Validate(context.Background(), result.PlainKey)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, project.ID, claims.ProjectID)
}

func TestApiKey_Validate_InvalidKey(t *testing.T) {
	_, svc, _, _ := setupApiKeyTest(t)

	_, err := svc.Validate(context.Background(), "ck_0000000000000000000000000000000000000000000000000000000000000000")
	assert.Error(t, err)
}

func TestApiKey_Validate_RevokedKey(t *testing.T) {
	_, svc, project, user := setupApiKeyTest(t)

	result, err := svc.Create(context.Background(), project.ID, user.ID, "revoke-me", nil)
	require.NoError(t, err)

	err = svc.Delete(context.Background(), result.ApiKey.ID, project.ID)
	require.NoError(t, err)

	_, err = svc.Validate(context.Background(), result.PlainKey)
	assert.Error(t, err, "revoked key should not validate")
}

func TestApiKey_List_Scoped(t *testing.T) {
	_, svc, project, user := setupApiKeyTest(t)

	_, err := svc.Create(context.Background(), project.ID, user.ID, "key-1", nil)
	require.NoError(t, err)
	_, err = svc.Create(context.Background(), project.ID, user.ID, "key-2", nil)
	require.NoError(t, err)

	keys, err := svc.List(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Len(t, keys, 2)

	// Verify no plaintext keys leak
	for _, k := range keys {
		assert.NotContains(t, k.KeyPrefix, "ck_", "list should never return plaintext")
		assert.Len(t, k.KeyPrefix, 8, "prefix should be 8 hex chars")
	}
}

func TestApiKey_List_Empty(t *testing.T) {
	_, svc, project, _ := setupApiKeyTest(t)

	keys, err := svc.List(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestApiKey_Delete_WrongProject(t *testing.T) {
	pool, svc, project, user := setupApiKeyTest(t)

	admin := testhelper.CreateTestAdmin(t, pool)
	otherProject := testhelper.CreateTestProject(t, pool, admin.ID)

	result, err := svc.Create(context.Background(), project.ID, user.ID, "wrong-project-delete", nil)
	require.NoError(t, err)

	// Delete with wrong project ID should not delete the key
	err = svc.Delete(context.Background(), result.ApiKey.ID, otherProject.ID)
	require.NoError(t, err)

	// Key should still validate
	_, err = svc.Validate(context.Background(), result.PlainKey)
	require.NoError(t, err, "key should still be valid after delete with wrong project")
}
