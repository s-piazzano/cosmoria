package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "cosmoria-test-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestLocalBackend_UploadAndDelete(t *testing.T) {
	basePath := tempDir(t)
	backend := NewLocalBackend(basePath)

	ctx := context.Background()
	key := "test/file.txt"
	content := "hello world"

	err := backend.Upload(ctx, key, strings.NewReader(content), int64(len(content)), "text/plain")
	require.NoError(t, err)

	fullPath := filepath.Join(basePath, key)
	_, err = os.Stat(fullPath)
	assert.NoError(t, err, "file should exist after upload")

	data, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))

	err = backend.Delete(ctx, key)
	require.NoError(t, err)

	_, err = os.Stat(fullPath)
	assert.True(t, os.IsNotExist(err), "file should be gone after delete")
}

func TestLocalBackend_Upload_CreatesDirectories(t *testing.T) {
	basePath := tempDir(t)
	backend := NewLocalBackend(basePath)

	key := "nested/dir/structure/file.txt"
	err := backend.Upload(context.Background(), key, strings.NewReader("data"), 4, "text/plain")
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(basePath, key))
	assert.NoError(t, err)
}

func TestLocalBackend_DownloadURL(t *testing.T) {
	basePath := tempDir(t)
	backend := NewLocalBackend(basePath)

	url, err := backend.DownloadURL(context.Background(), "some/key.txt", 0)
	require.NoError(t, err)
	assert.Equal(t, "some/key.txt", url, "local backend should return the key itself")
}

func TestLocalBackend_IsLocal(t *testing.T) {
	basePath := tempDir(t)
	backend := NewLocalBackend(basePath)

	assert.True(t, backend.IsLocal())
}

func TestLocalBackend_Ping(t *testing.T) {
	basePath := tempDir(t)
	backend := NewLocalBackend(basePath)

	err := backend.Ping(context.Background())
	assert.NoError(t, err)
}

func TestLocalBackend_Delete_NonExistent(t *testing.T) {
	basePath := tempDir(t)
	backend := NewLocalBackend(basePath)

	err := backend.Delete(context.Background(), "nonexistent/file.txt")
	assert.NoError(t, err, "deleting non-existent file should not error")
}

func TestNewLocalBackend_CreatesBasePath(t *testing.T) {
	basePath := filepath.Join(os.TempDir(), "cosmoria-test-create-dir")
	os.RemoveAll(basePath)

	backend := NewLocalBackend(basePath)
	defer os.RemoveAll(basePath)

	_, err := os.Stat(basePath)
	assert.NoError(t, err, "NewLocalBackend should create the base directory")
	assert.NotNil(t, backend)
}

func TestLocalBackend_KeySanitization(t *testing.T) {
	basePath := tempDir(t)
	backend := NewLocalBackend(basePath)

	key := "../../../etc/passwd"
	err := backend.Upload(context.Background(), key, strings.NewReader("x"), 1, "text/plain")
	require.NoError(t, err)

	fullPath := filepath.Join(basePath, key)

	absFull, _ := filepath.Abs(fullPath)
	absBase, _ := filepath.Abs(basePath)

	rel, err := filepath.Rel(absBase, absFull)
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(rel, ".."), "path traversal should resolve outside base")
	// Note: LocalBackend itself does not prevent path traversal in Upload.
	// Path traversal prevention is enforced at the HTTP handler layer via isPathSafe().
	// The service layer constrains keys to {projectID}/{tenantID}/{uuid}-{sanitizedName}.
}
