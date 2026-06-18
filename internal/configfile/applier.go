package configfile

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/collections"
	"github.com/s-piazzano/cosmoria/internal/core"
	"github.com/s-piazzano/cosmoria/internal/rbac"
	"github.com/s-piazzano/cosmoria/internal/tenant"
	"golang.org/x/crypto/bcrypt"
)

func ApplyIfPresent(pool *pgxpool.Pool, cfg *core.Config) error {
	cosmoriaCfg, err := Parse("cosmoria.yaml")
	if err != nil {
		return err
	}
	if cosmoriaCfg == nil {
		return nil
	}

	log.Println("config: applying cosmoria.yaml")
	return Apply(context.Background(), pool, cosmoriaCfg)
}

func Apply(ctx context.Context, pool *pgxpool.Pool, cfg *Config) error {
	adminSvc := adminauth.NewService(pool, nil)
	tenantSvc := tenant.NewService(pool)
	collSvc := collections.NewService(pool)
	rbacSvc := rbac.NewService(pool)

	adminID, err := ensureAdmin(ctx, pool)
	if err != nil {
		return fmt.Errorf("admin: %w", err)
	}

	projectID, err := ensureProject(ctx, pool, adminSvc, adminID, cfg.Project)
	if err != nil {
		return fmt.Errorf("project: %w", err)
	}

	for _, tc := range cfg.Tenants {
		if _, err := ensureTenant(ctx, tenantSvc, projectID, tc.Name); err != nil {
			return fmt.Errorf("tenant %q: %w", tc.Name, err)
		}
	}

	for _, cc := range cfg.Collections {
		if _, err := ensureCollection(ctx, collSvc, projectID, cc.Name, cc.Schema); err != nil {
			return fmt.Errorf("collection %q: %w", cc.Name, err)
		}
	}

	for _, rc := range cfg.Roles {
		if err := ensureRole(ctx, rbacSvc, pool, projectID, rc); err != nil {
			return fmt.Errorf("role %q: %w", rc.Name, err)
		}
	}

	return nil
}

func ensureAdmin(ctx context.Context, pool *pgxpool.Pool) (string, error) {
	var count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count); err != nil {
		return "", fmt.Errorf("check admin count: %w", err)
	}
	if count > 0 {
		var id string
		err := pool.QueryRow(ctx, `SELECT id FROM admin_users WHERE role = 'super_admin' ORDER BY created_at LIMIT 1`).Scan(&id)
		return id, err
	}

	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return "", fmt.Errorf("ADMIN_EMAIL and ADMIN_PASSWORD must be set (no admin users exist)")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	var id string
	err = pool.QueryRow(ctx,
		`INSERT INTO admin_users (email, password_hash, role) VALUES ($1, $2, 'super_admin') RETURNING id`,
		email, string(hash),
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create admin: %w", err)
	}

	return id, nil
}

func ensureProject(ctx context.Context, pool *pgxpool.Pool, svc *adminauth.Service, adminID, name string) (string, error) {
	var id string
	err := pool.QueryRow(ctx,
		`SELECT id FROM projects WHERE name = $1`, name,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != pgx.ErrNoRows {
		return "", fmt.Errorf("lookup project: %w", err)
	}

	p, err := svc.CreateProject(ctx, adminID, name, false)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	return p.ID, nil
}

func ensureTenant(ctx context.Context, svc *tenant.Service, projectID, name string) (string, error) {
	tenants, err := svc.ListTenants(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("list: %w", err)
	}
	for _, t := range tenants {
		if t.Name == name {
			return t.ID, nil
		}
	}

	t, err := svc.CreateTenant(ctx, projectID, name)
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	return t.ID, nil
}

func ensureCollection(ctx context.Context, svc *collections.Service, projectID, name string, sc SchemaConfig) (string, error) {
	cols, err := svc.ListCollections(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("list: %w", err)
	}
	for _, c := range cols {
		if c.Name == name {
			_, err := svc.UpdateCollectionSchema(ctx, c.ID, projectID, toSchema(sc))
			return c.ID, err
		}
	}

	c, err := svc.CreateCollection(ctx, projectID, name, toSchema(sc))
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	return c.ID, nil
}

func ensureRole(ctx context.Context, svc *rbac.Service, pool *pgxpool.Pool, projectID string, rc RoleConfig) error {
	roles, err := svc.ListRoles(ctx, projectID)
	if err != nil {
		return err
	}

	var roleID string
	for _, r := range roles {
		if r.Name == rc.Name {
			roleID = r.ID
			break
		}
	}

	if roleID == "" {
		r, err := svc.CreateRole(ctx, projectID, rc.Name)
		if err != nil {
			return fmt.Errorf("create role: %w", err)
		}
		roleID = r.ID
	}

	for _, pc := range rc.Permissions {
		if _, err := svc.SetPermission(ctx, roleID, pc.Resource, pc.Action); err != nil {
			return fmt.Errorf("set permission %s/%s: %w", pc.Resource, pc.Action, err)
		}
	}

	return nil
}

func toSchema(sc SchemaConfig) collections.Schema {
	fields := make([]collections.Field, len(sc.Fields))
	for i, f := range sc.Fields {
		fields[i] = collections.Field{Name: f.Name, Type: f.Type, Required: f.Required}
	}
	return collections.Schema{Fields: fields}
}
