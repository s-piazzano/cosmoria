package adminauth

import (
	"context"
	"fmt"
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
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	AdminOwnerID  string    `json:"admin_owner_id"`
	JWTExpiry     *int64    `json:"jwt_expiry,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
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

func (s *Service) Setup(ctx context.Context, email, password, projectName string) (*AuthResult, *Project, error) {
	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&count)
	if err != nil {
		return nil, nil, fmt.Errorf("adminauth: check count: %w", err)
	}
	if count > 0 {
		return nil, nil, fmt.Errorf("adminauth: already initialized")
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("adminauth: hash password: %w", err)
	}

	var adminID string
	err = s.pool.QueryRow(ctx,
		`INSERT INTO admin_users (email, password_hash, role) VALUES ($1, $2, 'super_admin') RETURNING id`,
		email, string(hash),
	).Scan(&adminID)
	if err != nil {
		return nil, nil, fmt.Errorf("adminauth: create admin: %w", err)
	}

	var project Project
	err = s.pool.QueryRow(ctx,
		`INSERT INTO projects (name, admin_owner_id) VALUES ($1, $2) RETURNING id, name, admin_owner_id, created_at`,
		projectName, adminID,
	).Scan(&project.ID, &project.Name, &project.AdminOwnerID, &project.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("adminauth: create project: %w", err)
	}

	token, err := GenerateToken(AdminClaims{
		AdminUserID: adminID,
		Role:        "super_admin",
	}, s.cfg.AdminJWTSecret, s.cfg.AdminJWTExpiry)
	if err != nil {
		return nil, nil, err
	}

	return &AuthResult{
		Token: token,
		Admin: AdminUser{
			ID:    adminID,
			Email: email,
			Role:  "super_admin",
		},
	}, &project, nil
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

func (s *Service) CreateProject(ctx context.Context, adminUserID, name string) (*Project, error) {
	var project Project
	err := s.pool.QueryRow(ctx,
		`INSERT INTO projects (name, admin_owner_id) VALUES ($1, $2) RETURNING id, name, admin_owner_id, created_at`,
		name, adminUserID,
	).Scan(&project.ID, &project.Name, &project.AdminOwnerID, &project.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("adminauth: create project: %w", err)
	}
	return &project, nil
}

func (s *Service) ListAccessibleProjects(ctx context.Context, adminUserID, role string) ([]ProjectWithRole, error) {
	var err error
	r, err := s.pool.Query(ctx, /* query */
		`SELECT id, name, admin_owner_id, created_at, 'super_admin' FROM projects ORDER BY created_at`,
	)
	if role != "super_admin" {
		r, err = s.pool.Query(ctx,
			`SELECT p.id, p.name, p.admin_owner_id, p.created_at, apr.role
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
		if err := r.Scan(&p.ID, &p.Name, &p.AdminOwnerID, &p.CreatedAt, &p.Role); err != nil {
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

type AdminProjectRole struct {
	AdminUserID string    `json:"admin_user_id"`
	ProjectID   string    `json:"project_id"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	Email       string    `json:"email,omitempty"`
}
