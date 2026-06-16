package testhelper

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/auth"
)

type TestAdmin struct {
	ID       string
	Email    string
	Password string
}

type TestProject struct {
	ID   string
	Name string
}

type TestUser struct {
	ID       string
	Email    string
	Password string
}

type TestTenant struct {
	ID   string
	Name string
}

var adminSeq int64

func CreateTestAdmin(t testing.TB, pool *pgxpool.Pool) *TestAdmin {
	t.Helper()

	n := atomic.AddInt64(&adminSeq, 1)
	a := &TestAdmin{
		Email:    fmt.Sprintf("admin%d@test.com", n),
		Password: "adminpass123",
	}

	hash, err := auth.HashPassword(a.Password)
	if err != nil {
		t.Fatalf("testhelper: hash admin password: %v", err)
	}

	err = pool.QueryRow(context.Background(),
		`INSERT INTO admin_users (email, password_hash, role)
		 VALUES ($1, $2, 'super_admin')
		 RETURNING id`,
		a.Email, string(hash),
	).Scan(&a.ID)
	if err != nil {
		t.Fatalf("testhelper: create admin: %v", err)
	}

	return a
}

func CreateTestProject(t testing.TB, pool *pgxpool.Pool, adminID string) *TestProject {
	t.Helper()

	p := &TestProject{Name: "Test Project"}

	err := pool.QueryRow(context.Background(),
		`INSERT INTO projects (name, admin_owner_id)
		 VALUES ($1, $2)
		 RETURNING id`,
		p.Name, adminID,
	).Scan(&p.ID)
	if err != nil {
		t.Fatalf("testhelper: create project: %v", err)
	}

	return p
}

var userSeq int64

func CreateTestUser(t testing.TB, pool *pgxpool.Pool) *TestUser {
	t.Helper()

	n := atomic.AddInt64(&userSeq, 1)
	u := &TestUser{
		Email:    fmt.Sprintf("user%d@test.com", n),
		Password: "userpass123",
	}

	hash, err := auth.HashPassword(u.Password)
	if err != nil {
		t.Fatalf("testhelper: hash user password: %v", err)
	}

	err = pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash)
		 VALUES ($1, $2)
		 RETURNING id`,
		u.Email, string(hash),
	).Scan(&u.ID)
	if err != nil {
		t.Fatalf("testhelper: create user: %v", err)
	}

	return u
}

func CreateTestTenant(t testing.TB, pool *pgxpool.Pool, projectID string) *TestTenant {
	t.Helper()

	tn := &TestTenant{Name: "Test Tenant"}

	err := pool.QueryRow(context.Background(),
		`INSERT INTO tenants (project_id, name)
		 VALUES ($1, $2)
		 RETURNING id`,
		projectID, tn.Name,
	).Scan(&tn.ID)
	if err != nil {
		t.Fatalf("testhelper: create tenant: %v", err)
	}

	return tn
}
