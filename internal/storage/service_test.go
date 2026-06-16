package storage_test

import (
	"context"
	"crypto/rand"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/storage"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

type mockBackend struct {
	storage.StorageBackend
	uploads   []string
	deletes   []string
	localURL  string
}

func (m *mockBackend) Upload(_ context.Context, key string, _ io.Reader, _ int64, _ string) error {
	m.uploads = append(m.uploads, key)
	return nil
}

func (m *mockBackend) DownloadURL(_ context.Context, key string, _ time.Duration) (string, error) {
	return m.localURL + key, nil
}

func (m *mockBackend) Delete(_ context.Context, key string) error {
	m.deletes = append(m.deletes, key)
	return nil
}

func (m *mockBackend) Ping(_ context.Context) error { return nil }

func (m *mockBackend) IsLocal() bool { return true }

func setupStorageTest(t *testing.T) (*pgxpool.Pool, *storage.Service, *mockBackend, *testhelper.TestProject, *testhelper.TestTenant, *testhelper.TestUser) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	backend := &mockBackend{localURL: "/api/download/"}
	svc := storage.NewService(pool, backend)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	tenant := testhelper.CreateTestTenant(t, pool, project.ID)
	user := testhelper.CreateTestUser(t, pool)

	return pool, svc, backend, project, tenant, user
}

func TestStorage_Upload(t *testing.T) {
	_, svc, backend, project, tenant, user := setupStorageTest(t)

	content := "test content"
	reader := strings.NewReader(content)

	f, err := svc.Upload(context.Background(), project.ID, tenant.ID, "test.txt", "text/plain", reader, int64(len(content)), user.ID)
	require.NoError(t, err)
	assert.NotEmpty(t, f.ID)
	assert.Equal(t, project.ID, f.ProjectID)
	assert.Equal(t, tenant.ID, f.TenantID)
	assert.Equal(t, "test.txt", f.Filename)
	assert.Equal(t, user.ID, f.UploadedBy)

	require.Len(t, backend.uploads, 1)
	assert.Contains(t, backend.uploads[0], project.ID+"/"+tenant.ID+"/")
	assert.Contains(t, backend.uploads[0], "-test.txt")
}

func TestStorage_GetMeta(t *testing.T) {
	_, svc, _, project, tenant, user := setupStorageTest(t)

	created, err := svc.Upload(context.Background(), project.ID, tenant.ID, "meta.txt", "text/plain", strings.NewReader("data"), 4, user.ID)
	require.NoError(t, err)

	meta, err := svc.GetMeta(context.Background(), project.ID, tenant.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, meta.ID)
	assert.Equal(t, "meta.txt", meta.Filename)
}

func TestStorage_GetByID(t *testing.T) {
	_, svc, _, project, tenant, user := setupStorageTest(t)

	created, err := svc.Upload(context.Background(), project.ID, tenant.ID, "get.txt", "text/plain", strings.NewReader("data"), 4, user.ID)
	require.NoError(t, err)

	f, err := svc.GetByID(context.Background(), project.ID, tenant.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, f.ID)
	assert.Contains(t, f.PresignedURL, "/api/projects/"+project.ID+"/tenants/"+tenant.ID+"/files/"+created.ID+"/download")
}

func TestStorage_List(t *testing.T) {
	_, svc, _, project, tenant, user := setupStorageTest(t)

	for range 5 {
		_, err := svc.Upload(context.Background(), project.ID, tenant.ID, "file.txt", "text/plain", strings.NewReader("x"), 1, user.ID)
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}

	list, next, err := svc.List(context.Background(), project.ID, tenant.ID, "", 3)
	require.NoError(t, err)
	assert.Len(t, list, 3)
	assert.NotEmpty(t, next)

	list2, next, err := svc.List(context.Background(), project.ID, tenant.ID, next, 3)
	require.NoError(t, err)
	assert.Len(t, list2, 1)
	assert.Empty(t, next)
}

func TestStorage_Delete(t *testing.T) {
	_, svc, backend, project, tenant, user := setupStorageTest(t)

	created, err := svc.Upload(context.Background(), project.ID, tenant.ID, "del.txt", "text/plain", strings.NewReader("data"), 4, user.ID)
	require.NoError(t, err)

	err = svc.Delete(context.Background(), project.ID, tenant.ID, created.ID)
	require.NoError(t, err)

	require.Len(t, backend.deletes, 1)
	assert.Equal(t, created.S3Key, backend.deletes[0])

	_, err = svc.GetMeta(context.Background(), project.ID, tenant.ID, created.ID)
	assert.Error(t, err, "metadata should be removed after delete")
}

func TestStorage_Delete_NotFound(t *testing.T) {
	_, svc, _, project, tenant, _ := setupStorageTest(t)

	err := svc.Delete(context.Background(), project.ID, tenant.ID, "00000000-0000-0000-0000-000000000000")
	assert.Error(t, err)
}

func TestStorage_Upload_SanitizesFilename(t *testing.T) {
	_, svc, backend, project, tenant, user := setupStorageTest(t)

	content := "test"
	_, err := svc.Upload(context.Background(), project.ID, tenant.ID, "../../etc/passwd", "text/plain", strings.NewReader(content), int64(len(content)), user.ID)
	require.NoError(t, err)

	require.Len(t, backend.uploads, 1)
	// The sanitized filename should NOT contain "../"
	assert.NotContains(t, backend.uploads[0], "../")
	assert.Contains(t, backend.uploads[0], "-.._.._etc_passwd")
}

func TestStorage_PaginationDefaults(t *testing.T) {
	_, svc, _, project, tenant, _ := setupStorageTest(t)

	list, next, err := svc.List(context.Background(), project.ID, tenant.ID, "", 0)
	require.NoError(t, err)
	assert.Empty(t, list)
	assert.Empty(t, next)
}

func TestStorage_Upload_RandomContent(t *testing.T) {
	_, svc, _, project, tenant, user := setupStorageTest(t)

	data := make([]byte, 256)
	_, err := rand.Read(data)
	require.NoError(t, err)

	reader := strings.NewReader(string(data))
	f, err := svc.Upload(context.Background(), project.ID, tenant.ID, "random.bin", "application/octet-stream", reader, int64(len(data)), user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(256), f.Size)
}
