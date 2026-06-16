package records_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/collections"
	"github.com/s-piazzano/cosmoria/internal/records"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

type recordsFixture struct {
	pool          *pgxpool.Pool
	svc           *records.Service
	colls         *collections.Service
	project       *testhelper.TestProject
	tenant        *testhelper.TestTenant
	collectionID  string
	schemaFields  []collections.Field
}

func setupRecordsTest(t *testing.T) *recordsFixture {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	colls := collections.NewService(pool)
	svc := records.NewService(pool, colls)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	tenant := testhelper.CreateTestTenant(t, pool, project.ID)

	schema := collections.Schema{
		Fields: []collections.Field{
			{Name: "title", Type: "string", Required: true},
			{Name: "count", Type: "number"},
		},
	}
	coll, err := colls.CreateCollection(context.Background(), project.ID, "posts", schema)
	require.NoError(t, err)

	return &recordsFixture{
		pool:         pool,
		svc:          svc,
		colls:        colls,
		project:      project,
		tenant:       tenant,
		collectionID: coll.ID,
		schemaFields: schema.Fields,
	}
}

func TestRecords_CreateAndGet(t *testing.T) {
	f := setupRecordsTest(t)

	record, err := f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
		"title": "Hello World",
		"count": float64(42),
	})
	require.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.Equal(t, f.project.ID, record.ProjectID)
	assert.Equal(t, f.tenant.ID, record.TenantID)
	assert.Equal(t, f.collectionID, record.CollectionID)
	assert.Equal(t, "Hello World", record.Data["title"])

	got, err := f.svc.GetRecord(context.Background(), record.ID, f.project.ID, f.tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, record.ID, got.ID)
	assert.Equal(t, "Hello World", got.Data["title"])
}

func TestRecords_Create_InvalidSchema(t *testing.T) {
	f := setupRecordsTest(t)

	_, err := f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
		"count": 42,
	})
	assert.Error(t, err, "missing required field 'title'")

	_, err = f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
		"title": "Test",
		"count": "not-a-number",
	})
	assert.Error(t, err, "type mismatch on 'count'")
}

func TestRecords_Update(t *testing.T) {
	f := setupRecordsTest(t)

	record, err := f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
		"title": "Original",
	})
	require.NoError(t, err)

	updated, err := f.svc.UpdateRecord(context.Background(), record.ID, f.project.ID, f.tenant.ID, map[string]any{
		"title": "Updated",
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.Data["title"])
}

func TestRecords_Update_InvalidData(t *testing.T) {
	f := setupRecordsTest(t)

	record, err := f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
		"title": "Original",
	})
	require.NoError(t, err)

	_, err = f.svc.UpdateRecord(context.Background(), record.ID, f.project.ID, f.tenant.ID, map[string]any{
		"title": 123,
	})
	assert.Error(t, err, "type mismatch should fail update")
}

func TestRecords_Delete(t *testing.T) {
	f := setupRecordsTest(t)

	record, err := f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
		"title": "Delete Me",
	})
	require.NoError(t, err)

	err = f.svc.DeleteRecord(context.Background(), record.ID, f.project.ID, f.tenant.ID)
	require.NoError(t, err)

	_, err = f.svc.GetRecord(context.Background(), record.ID, f.project.ID, f.tenant.ID)
	assert.Error(t, err)
}

func TestRecords_List_Pagination(t *testing.T) {
	f := setupRecordsTest(t)

	for i := range 10 {
		_, err := f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
			"title": fmt.Sprintf("Record %d", i),
		})
		require.NoError(t, err)
		time.Sleep(time.Millisecond)
	}

	results, nextCursor, err := f.svc.ListRecords(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, "", 3)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.NotEmpty(t, nextCursor, "should have more pages")

	results, nextCursor, err = f.svc.ListRecords(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, nextCursor, 3)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestRecords_List_DefaultLimit(t *testing.T) {
	f := setupRecordsTest(t)

	results, _, err := f.svc.ListRecords(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, "", 0)
	require.NoError(t, err)
	assert.Len(t, results, 0, "no records yet, but should not error with default limit")
}

func TestRecords_TenantIsolation(t *testing.T) {
	f := setupRecordsTest(t)

	record, err := f.svc.CreateRecord(context.Background(), f.project.ID, f.tenant.ID, f.collectionID, map[string]any{
		"title": "Secret",
	})
	require.NoError(t, err)

	otherTenant := testhelper.CreateTestTenant(t, f.pool, f.project.ID)

	_, err = f.svc.GetRecord(context.Background(), record.ID, f.project.ID, otherTenant.ID)
	assert.Error(t, err, "should not access record from another tenant")
}
