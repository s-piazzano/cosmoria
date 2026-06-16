package audit_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/s-piazzano/cosmoria/internal/audit"
	"github.com/s-piazzano/cosmoria/internal/testhelper"
)

func setupAuditTest(t *testing.T) (*pgxpool.Pool, *audit.Service, *audit.Logger, *testhelper.TestProject, string) {
	t.Helper()

	pool := testhelper.NewTestDB(t)
	svc := audit.NewService(pool)
	logger := audit.NewLogger(pool)

	admin := testhelper.CreateTestAdmin(t, pool)
	project := testhelper.CreateTestProject(t, pool, admin.ID)
	user := testhelper.CreateTestUser(t, pool)

	return pool, svc, logger, project, user.ID
}

func TestAudit_LogAndList(t *testing.T) {
	_, svc, logger, project, userID := setupAuditTest(t)

	logger.Log(context.Background(), project.ID, userID, "create", "tenants", nil, json.RawMessage(`{"name":"acme"}`), "127.0.0.1")

	time.Sleep(500 * time.Millisecond)

	entries, nextCursor, err := svc.List(context.Background(), project.ID, "", 10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "create", entries[0].Action)
	assert.Equal(t, "tenants", entries[0].Resource)
	assert.Equal(t, userID, entries[0].UserID)
	assert.Equal(t, "127.0.0.1", entries[0].IPAddress)
	assert.Empty(t, nextCursor)
}

func TestAudit_List_ScopedByProject(t *testing.T) {
	pool, svc, logger, project, userID := setupAuditTest(t)

	otherAdmin := testhelper.CreateTestAdmin(t, pool)
	otherProject := testhelper.CreateTestProject(t, pool, otherAdmin.ID)

	logger.Log(context.Background(), project.ID, userID, "delete", "records", nil, json.RawMessage(`{"reason":"test"}`), "10.0.0.1")
	logger.Log(context.Background(), otherProject.ID, userID, "create", "records", nil, nil, "10.0.0.2")

	time.Sleep(500 * time.Millisecond)

	entries, _, err := svc.List(context.Background(), project.ID, "", 10)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "should only list audit entries for the project")
}

func TestAudit_List_Pagination(t *testing.T) {
	_, svc, logger, project, userID := setupAuditTest(t)

	for range 5 {
		logger.Log(context.Background(), project.ID, userID, "create", "records", nil, nil, "10.0.0.1")
		time.Sleep(time.Millisecond)
	}

	require.Eventually(t, func() bool {
		entries, _, _ := svc.List(context.Background(), project.ID, "", 10)
		return len(entries) == 5
	}, 5*time.Second, 100*time.Millisecond, "audit log entries should appear")

	entries, nextCursor, err := svc.List(context.Background(), project.ID, "", 3)
	require.NoError(t, err)
	assert.Len(t, entries, 3)
	assert.NotEmpty(t, nextCursor)

	entries2, nextCursor, err := svc.List(context.Background(), project.ID, nextCursor, 3)
	require.NoError(t, err)
	assert.Len(t, entries2, 1)
	assert.Empty(t, nextCursor, "should be last page")
}

func TestAudit_List_DefaultLimit(t *testing.T) {
	_, svc, logger, project, userID := setupAuditTest(t)

	for range 60 {
		logger.Log(context.Background(), project.ID, userID, "read", "files", nil, nil, "10.0.0.1")
	}

	time.Sleep(2000 * time.Millisecond)

	entries, _, err := svc.List(context.Background(), project.ID, "", 0)
	require.NoError(t, err)
	assert.Len(t, entries, 50, "default limit should be 50")
}

func TestAudit_List_MaxLimit(t *testing.T) {
	_, svc, logger, project, userID := setupAuditTest(t)

	for range 120 {
		logger.Log(context.Background(), project.ID, userID, "read", "files", nil, nil, "10.0.0.1")
	}

	time.Sleep(3000 * time.Millisecond)

	entries, _, err := svc.List(context.Background(), project.ID, "", 200)
	require.NoError(t, err)
	assert.Len(t, entries, 100, "max limit should be 100")
}

func TestAudit_Log_DoesNotBlock(t *testing.T) {
	_, _, logger, project, userID := setupAuditTest(t)

	done := make(chan bool)
	go func() {
		logger.Log(context.Background(), project.ID, userID, "login", "auth", nil, nil, "10.0.0.1")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Log should not block")
	}
}

func stringPtr(s string) *string { return &s }
