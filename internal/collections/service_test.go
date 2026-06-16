package collections_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/collections"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupCollectionsTest(t *testing.T) (*pgxpool.Pool, *collections.Service, *testhelper.TestProject) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	svc := collections.NewService(pool)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)

	return pool, svc, project
}

func TestCollections_CRUD(t *testing.T) {
	_, svc, project := setupCollectionsTest(t)

	schema := collections.Schema{
		Fields: []collections.Field{
			{Name: "title", Type: "string", Required: true},
			{Name: "body", Type: "string"},
		},
	}

	created, err := svc.CreateCollection(context.Background(), project.ID, "posts", schema)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "posts", created.Name)
	assert.Equal(t, project.ID, created.ProjectID)
	assert.Equal(t, 2, len(created.Schema.Fields))

	got, err := svc.GetCollection(context.Background(), created.ID, project.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "posts", got.Name)

	updatedSchema := collections.Schema{
		Fields: []collections.Field{
			{Name: "title", Type: "string", Required: true},
			{Name: "body", Type: "string"},
			{Name: "tags", Type: "string"},
		},
	}
	updated, err := svc.UpdateCollectionSchema(context.Background(), created.ID, project.ID, updatedSchema)
	require.NoError(t, err)
	assert.Len(t, updated.Schema.Fields, 3)

	err = svc.DeleteCollection(context.Background(), created.ID, project.ID)
	require.NoError(t, err)

	_, err = svc.GetCollection(context.Background(), created.ID, project.ID)
	assert.Error(t, err)
}

func TestCollections_List(t *testing.T) {
	_, svc, project := setupCollectionsTest(t)

	_, err := svc.CreateCollection(context.Background(), project.ID, "first", collections.Schema{})
	require.NoError(t, err)
	_, err = svc.CreateCollection(context.Background(), project.ID, "second", collections.Schema{})
	require.NoError(t, err)

	list, err := svc.ListCollections(context.Background(), project.ID)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestCollections_ProjectIsolation(t *testing.T) {
	pool, svc, projectA := setupCollectionsTest(t)

	admin := testhelper.CreateTestAdmin(t, pool)
	projectB := testhelper.CreateTestProject(t, pool, admin.ID)

	coll, err := svc.CreateCollection(context.Background(), projectA.ID, "secret", collections.Schema{})
	require.NoError(t, err)

	_, err = svc.GetCollection(context.Background(), coll.ID, projectB.ID)
	assert.Error(t, err, "should not find collection from different project")
}

func TestValidateData_RequiredFields(t *testing.T) {
	schema := collections.Schema{
		Fields: []collections.Field{
			{Name: "title", Type: "string", Required: true},
			{Name: "body", Type: "string"},
		},
	}

	err := collections.ValidateData(map[string]any{"title": "Hello"}, schema)
	assert.NoError(t, err)

	err = collections.ValidateData(map[string]any{}, schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title")

	err = collections.ValidateData(map[string]any{"title": nil}, schema)
	assert.Error(t, err)
}

func TestValidateData_TypeChecking(t *testing.T) {
	schema := collections.Schema{
		Fields: []collections.Field{
			{Name: "name", Type: "string"},
			{Name: "age", Type: "number"},
			{Name: "active", Type: "boolean"},
		},
	}

	err := collections.ValidateData(map[string]any{"name": "Alice", "age": float64(30), "active": true}, schema)
	assert.NoError(t, err)

	err = collections.ValidateData(map[string]any{"name": 42, "age": float64(30), "active": true}, schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name")

	err = collections.ValidateData(map[string]any{"name": "Alice", "age": "thirty", "active": true}, schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "age")

	err = collections.ValidateData(map[string]any{"name": "Alice", "age": float64(30), "active": "yes"}, schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active")
}

func TestValidateData_UnsupportedType(t *testing.T) {
	schema := collections.Schema{
		Fields: []collections.Field{
			{Name: "data", Type: "binary"},
		},
	}

	err := collections.ValidateData(map[string]any{"data": "bytes"}, schema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

func TestValidateData_EmptySchema(t *testing.T) {
	schema := collections.Schema{}
	err := collections.ValidateData(map[string]any{"anything": "ok"}, schema)
	assert.NoError(t, err)
}
