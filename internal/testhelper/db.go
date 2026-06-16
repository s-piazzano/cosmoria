package testhelper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/db"
)

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

var (
	setupOnce   sync.Once
	migrateOnce sync.Once
	setupErr    error
)

func TestDBURL() string {
	if v := os.Getenv("COSMORIA_TEST_DB_URL"); v != "" {
		return v
	}
	return "postgres://localhost:5432/cosmoria_test?sslmode=disable"
}

func adminDBURL() string {
	if v := os.Getenv("COSMORIA_ADMIN_DB_URL"); v != "" {
		return v
	}
	return "postgres://localhost:5432/postgres?sslmode=disable"
}

func NewTestDB(t testing.TB) *pgxpool.Pool {
	t.Helper()

	setupOnce.Do(func() {
		adminPool, err := pgxpool.New(context.Background(), adminDBURL())
		if err != nil {
			setupErr = fmt.Errorf("testhelper: connect admin DB: %w", err)
			return
		}
		defer adminPool.Close()

		_, err = adminPool.Exec(context.Background(), `CREATE DATABASE cosmoria_test`)
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			setupErr = fmt.Errorf("testhelper: create test DB: %w", err)
			return
		}
	})

	migrateOnce.Do(func() {
		migratePool, err := pgxpool.New(context.Background(), TestDBURL())
		if err != nil {
			setupErr = fmt.Errorf("testhelper: connect test DB: %w", err)
			return
		}
		defer migratePool.Close()

		if err := db.Migrate(migratePool, filepath.Join(projectRoot(), "db", "migrations")); err != nil {
			setupErr = fmt.Errorf("testhelper: migrate: %w", err)
		}
	})

	if setupErr != nil {
		t.Fatalf("setup failed: %v", setupErr)
	}

	pool, err := pgxpool.New(context.Background(), TestDBURL())
	if err != nil {
		t.Fatalf("testhelper: connect test DB: %v", err)
	}

	// Truncate all tables BEFORE the test to ensure clean state
	truncateAll(t, pool)

	t.Cleanup(func() {
		truncateAll(t, pool)
		pool.Close()
	})

	return pool
}

func truncateAll(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	// Deterministic order to avoid deadlocks between concurrent tests
	tables := []string{
		"audit_logs",
		"user_project_roles",
		"project_role_permissions",
		"project_roles",
		"user_tenants",
		"records",
		"files",
		"collections",
		"tenants",
		"api_keys",
		"users",
		"admin_project_roles",
		"projects",
		"admin_users",
	}
	for _, table := range tables {
		_, err := pool.Exec(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("testhelper: truncate %s: %v", table, err)
		}
	}
}
