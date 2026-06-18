package adminauth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/s-piazzano/cosmoria/internal/core"
)

type AdminUser struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Project struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	AdminOwnerID        string    `json:"admin_owner_id"`
	JWTExpiry           *int64    `json:"jwt_expiry,omitempty"`
	MultitenancyEnabled bool      `json:"multitenancy_enabled"`
	CreatedAt           time.Time `json:"created_at"`
}

type AuthResult struct {
	Token string     `json:"token"`
	Admin AdminUser  `json:"admin"`
}

type ProjectWithRole struct {
	Project
	Role string `json:"role"`
}

type Service struct {
	pool *pgxpool.Pool
	cfg  *core.Config
}

func NewService(pool *pgxpool.Pool, cfg *core.Config) *Service {
	return &Service{pool: pool, cfg: cfg}
}

func (s *Service) NeedsSetup(ctx context.Context) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("adminauth: check needs setup: %w", err)
	}
	return count == 0, nil
}

func (s *Service) Setup(ctx context.Context, email, password string) (*AuthResult, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("adminauth: check count: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("adminauth: already initialized")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("adminauth: hash password: %w", err)
	}

	var adminID string
	err = s.pool.QueryRow(ctx,
		`INSERT INTO admin_users (email, password_hash, role) VALUES ($1, $2, 'super_admin') RETURNING id`,
		email, string(hash),
	).Scan(&adminID)
	if err != nil {
		return nil, fmt.Errorf("adminauth: create admin: %w", err)
	}

	token, err := GenerateToken(AdminClaims{
		AdminUserID: adminID,
		Role:        "super_admin",
	}, s.cfg.AdminJWTSecret, s.cfg.AdminJWTExpiry)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token: token,
		Admin: AdminUser{
			ID:    adminID,
			Email: email,
			Role:  "super_admin",
		},
	}, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	var id, passwordHash, role string
	err := s.pool.QueryRow(ctx,
		`SELECT id, password_hash, role FROM admin_users WHERE email = $1`,
		email,
	).Scan(&id, &passwordHash, &role)
	if err != nil {
		return nil, fmt.Errorf("adminauth: invalid credentials")
	}

	if !CheckPassword(password, []byte(passwordHash)) {
		return nil, fmt.Errorf("adminauth: invalid credentials")
	}

	token, err := GenerateToken(AdminClaims{
		AdminUserID: id,
		Role:        role,
	}, s.cfg.AdminJWTSecret, s.cfg.AdminJWTExpiry)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token: token,
		Admin: AdminUser{
			ID:    id,
			Email: email,
			Role:  role,
		},
	}, nil
}

func (s *Service) CreateProject(ctx context.Context, adminUserID, name string, multitenancyEnabled bool) (*Project, error) {
	slug, err := s.uniqueSlug(ctx, generateSlug(name))
	if err != nil {
		return nil, fmt.Errorf("adminauth: generate slug: %w", err)
	}

	var project Project
	err = s.pool.QueryRow(ctx,
		`INSERT INTO projects (name, slug, admin_owner_id, multitenancy_enabled) VALUES ($1, $2, $3, $4) RETURNING id, name, slug, admin_owner_id, multitenancy_enabled, created_at`,
		name, slug, adminUserID, multitenancyEnabled,
	).Scan(&project.ID, &project.Name, &project.Slug, &project.AdminOwnerID, &project.MultitenancyEnabled, &project.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adminauth: create project: %w", err)
	}
	return &project, nil
}

func (s *Service) ToggleMultitenancy(ctx context.Context, projectID string, enabled bool) error {
	if !enabled {
		var count int
		err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants WHERE project_id = $1`, projectID).Scan(&count)
		if err != nil {
			return fmt.Errorf("adminauth: check tenants: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("cannot disable multitenancy: project has %d tenant(s); delete all tenants first", count)
		}
	}

	_, err := s.pool.Exec(ctx,
		`UPDATE projects SET multitenancy_enabled = $1 WHERE id = $2`,
		enabled, projectID,
	)
	if err != nil {
		return fmt.Errorf("adminauth: toggle multitenancy: %w", err)
	}
	return nil
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == ' ' {
			return r
		}
		return -1
	}, slug)
	slug = strings.ReplaceAll(slug, " ", "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")
	if len(slug) > 60 {
		slug = slug[:60]
	}
	slug = strings.TrimRight(slug, "-")
	if slug == "" {
		slug = "untitled"
	}
	return slug
}

func (s *Service) uniqueSlug(ctx context.Context, base string) (string, error) {
	slug := base
	for i := 2; ; i++ {
		var exists bool
		err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE slug = $1)`, slug).Scan(&exists)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		if i == 2 {
			slug = base + "-2"
		} else {
			slug = fmt.Sprintf("%s-%d", base, i)
		}
	}
}

func (s *Service) ListAccessibleProjects(ctx context.Context, adminUserID, role string) ([]ProjectWithRole, error) {
	var err error
	r, err := s.pool.Query(ctx, /* query */
		`SELECT id, name, slug, admin_owner_id, multitenancy_enabled, created_at, 'super_admin' FROM projects ORDER BY created_at`,
	)
	if role != "super_admin" {
		r, err = s.pool.Query(ctx,
			`SELECT p.id, p.name, p.slug, p.admin_owner_id, p.multitenancy_enabled, p.created_at, apr.role
			 FROM projects p
			 JOIN admin_project_roles apr ON apr.project_id = p.id AND apr.admin_user_id = $1
			 ORDER BY p.created_at`, adminUserID)
	}
	if err != nil {
		return nil, fmt.Errorf("adminauth: list projects: %w", err)
	}
	defer r.Close()

	var projects []ProjectWithRole
	for r.Next() {
		var p ProjectWithRole
		if err := r.Scan(&p.ID, &p.Name, &p.Slug, &p.AdminOwnerID, &p.MultitenancyEnabled, &p.CreatedAt, &p.Role); err != nil {
			return nil, fmt.Errorf("adminauth: scan project: %w", err)
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *Service) GetAdminUser(ctx context.Context, adminUserID string) (*AdminUser, error) {
	var u AdminUser
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, role, created_at FROM admin_users WHERE id = $1`,
		adminUserID,
	).Scan(&u.ID, &u.Email, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adminauth: get admin: %w", err)
	}
	return &u, nil
}

func (s *Service) AssignRole(ctx context.Context, projectID, adminUserID, role string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO admin_project_roles (admin_user_id, project_id, role) VALUES ($1, $2, $3) ON CONFLICT (admin_user_id, project_id) DO UPDATE SET role = $3`,
		adminUserID, projectID, role,
	)
	if err != nil {
		return fmt.Errorf("adminauth: assign role: %w", err)
	}
	return nil
}

func (s *Service) RemoveRole(ctx context.Context, projectID, adminUserID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM admin_project_roles WHERE admin_user_id = $1 AND project_id = $2`,
		adminUserID, projectID,
	)
	if err != nil {
		return fmt.Errorf("adminauth: remove role: %w", err)
	}
	return nil
}

func (s *Service) ListRoles(ctx context.Context, projectID string) ([]AdminProjectRole, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT apr.admin_user_id, apr.project_id, apr.role, apr.created_at, au.email
		 FROM admin_project_roles apr
		 JOIN admin_users au ON au.id = apr.admin_user_id
		 WHERE apr.project_id = $1
		 ORDER BY au.email`, projectID)
	if err != nil {
		return nil, fmt.Errorf("adminauth: list roles: %w", err)
	}
	defer rows.Close()

	var roles []AdminProjectRole
	for rows.Next() {
		var r AdminProjectRole
		if err := rows.Scan(&r.AdminUserID, &r.ProjectID, &r.Role, &r.CreatedAt, &r.Email); err != nil {
			return nil, fmt.Errorf("adminauth: scan role: %w", err)
		}
		roles = append(roles, r)
	}
	return roles, nil
}

type OverviewStats struct {
	Tenants     int `json:"tenants"`
	Collections int `json:"collections"`
	Users       int `json:"users"`
	Records     int `json:"records"`
	Files       int `json:"files"`
	Roles       int `json:"roles"`
}

func (s *Service) GetOverviewStats(ctx context.Context, projectID string) (*OverviewStats, error) {
	var stats OverviewStats
	err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM tenants WHERE project_id = $1),
			(SELECT COUNT(*) FROM collections WHERE project_id = $1),
			(SELECT COUNT(*) FROM user_project_roles WHERE project_id = $1),
			(SELECT COUNT(*) FROM records WHERE project_id = $1),
			(SELECT COUNT(*) FROM files WHERE project_id = $1),
			(SELECT COUNT(*) FROM project_roles WHERE project_id = $1)
	`, projectID).Scan(&stats.Tenants, &stats.Collections, &stats.Users, &stats.Records, &stats.Files, &stats.Roles)
	if err != nil {
		return nil, fmt.Errorf("adminauth: get overview stats: %w", err)
	}
	return &stats, nil
}

func (s *Service) HasProjectAccess(ctx context.Context, adminUserID, projectID string) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin_project_roles WHERE admin_user_id = $1 AND project_id = $2`,
		adminUserID, projectID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("adminauth: check project access: %w", err)
	}
	return count > 0, nil
}

func (s *Service) GetProject(ctx context.Context, id string) (*Project, error) {
	var p Project
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, slug, admin_owner_id, jwt_expiry, multitenancy_enabled, created_at FROM projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.AdminOwnerID, &p.JWTExpiry, &p.MultitenancyEnabled, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adminauth: get project: %w", err)
	}
	return &p, nil
}

type UpdateProjectInput struct {
	Name                string
	JWTExpiry           *int64
	MultitenancyEnabled *bool
}

func (s *Service) UpdateProject(ctx context.Context, id string, input UpdateProjectInput) (*Project, error) {
	// Handle multitenancy guard separately
	if input.MultitenancyEnabled != nil && !*input.MultitenancyEnabled {
		var count int
		err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants WHERE project_id = $1`, id).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("adminauth: check tenants: %w", err)
		}
		if count > 0 {
			return nil, fmt.Errorf("cannot disable multitenancy: project has %d tenant(s); delete all tenants first", count)
		}
	}

	var p Project
	err := s.pool.QueryRow(ctx,
		`UPDATE projects SET
			name = COALESCE(NULLIF($1, ''), name),
			jwt_expiry = $2,
			multitenancy_enabled = COALESCE($3, multitenancy_enabled)
		 WHERE id = $4
		 RETURNING id, name, slug, admin_owner_id, jwt_expiry, multitenancy_enabled, created_at`,
		input.Name, input.JWTExpiry, input.MultitenancyEnabled, id,
	).Scan(&p.ID, &p.Name, &p.Slug, &p.AdminOwnerID, &p.JWTExpiry, &p.MultitenancyEnabled, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adminauth: update project: %w", err)
	}
	return &p, nil
}

func (s *Service) DeleteProject(ctx context.Context, id string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("adminauth: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `DELETE FROM audit_logs WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete audit: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM files WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete files: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM records WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete records: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM user_project_roles WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete user roles: %w", err)
	}
	_, err = tx.Exec(ctx,
		`DELETE FROM project_role_permissions WHERE role_id IN (SELECT id FROM project_roles WHERE project_id = $1)`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete role permissions: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM project_roles WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete roles: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM user_tenants WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete user tenants: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM tenants WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete tenants: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM collections WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete collections: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM api_keys WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete api keys: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM admin_project_roles WHERE project_id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete admin roles: %w", err)
	}
	_, err = tx.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("adminauth: delete project: %w", err)
	}

	return tx.Commit(ctx)
}

type AdminProjectRole struct {
	AdminUserID string    `json:"admin_user_id"`
	ProjectID   string    `json:"project_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	Email       string    `json:"email,omitempty"`
}
