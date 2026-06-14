package rbac

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func (s *Service) CreateRole(ctx context.Context, projectID, name string) (*Role, error) {
	var r Role
	err := s.pool.QueryRow(ctx,
		`INSERT INTO project_roles (project_id, name) VALUES ($1, $2) RETURNING id, project_id, name, created_at`,
		projectID, name,
	).Scan(&r.ID, &r.ProjectID, &r.Name, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("rbac: create role: %w", err)
	}
	return &r, nil
}

func (s *Service) ListRoles(ctx context.Context, projectID string) ([]RoleWithPermissions, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, name, created_at FROM project_roles WHERE project_id = $1 ORDER BY name`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("rbac: list roles: %w", err)
	}
	defer rows.Close()

	var results []RoleWithPermissions
	for rows.Next() {
		var r RoleWithPermissions
		if err := rows.Scan(&r.ID, &r.ProjectID, &r.Name, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("rbac: scan role: %w", err)
		}
		results = append(results, r)
	}

	for i := range results {
		perms, err := s.ListPermissions(ctx, results[i].ID)
		if err != nil {
			return nil, err
		}
		results[i].Permissions = perms
	}

	return results, nil
}

func (s *Service) DeleteRole(ctx context.Context, roleID, projectID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM project_roles WHERE id = $1 AND project_id = $2`,
		roleID, projectID,
	)
	if err != nil {
		return fmt.Errorf("rbac: delete role: %w", err)
	}
	return nil
}

func (s *Service) SetPermission(ctx context.Context, roleID, resource, action string) (*Permission, error) {
	var p Permission
	err := s.pool.QueryRow(ctx,
		`INSERT INTO project_role_permissions (role_id, resource, action)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (role_id, resource, action) DO UPDATE SET resource = EXCLUDED.resource
		 RETURNING id, role_id, resource, action`,
		roleID, resource, action,
	).Scan(&p.ID, &p.RoleID, &p.Resource, &p.Action)
	if err != nil {
		return nil, fmt.Errorf("rbac: set permission: %w", err)
	}
	return &p, nil
}

func (s *Service) RemovePermission(ctx context.Context, roleID, resource, action string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM project_role_permissions WHERE role_id = $1 AND resource = $2 AND action = $3`,
		roleID, resource, action,
	)
	if err != nil {
		return fmt.Errorf("rbac: remove permission: %w", err)
	}
	return nil
}

func (s *Service) ListPermissions(ctx context.Context, roleID string) ([]Permission, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, role_id, resource, action FROM project_role_permissions WHERE role_id = $1 ORDER BY resource, action`,
		roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("rbac: list permissions: %w", err)
	}
	defer rows.Close()

	var perms []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.ID, &p.RoleID, &p.Resource, &p.Action); err != nil {
			return nil, fmt.Errorf("rbac: scan permission: %w", err)
		}
		perms = append(perms, p)
	}
	return perms, nil
}

func (s *Service) AssignUserRole(ctx context.Context, userID, projectID, roleID string) (*UserProjectRole, error) {
	var upr UserProjectRole
	err := s.pool.QueryRow(ctx,
		`INSERT INTO user_project_roles (user_id, project_id, role_id) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, project_id) DO UPDATE SET role_id = EXCLUDED.role_id
		 RETURNING user_id, project_id, role_id, created_at`,
		userID, projectID, roleID,
	).Scan(&upr.UserID, &upr.ProjectID, &upr.RoleID, &upr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("rbac: assign user role: %w", err)
	}
	return &upr, nil
}

func (s *Service) GetUserRole(ctx context.Context, userID, projectID string) (*UserProjectRole, error) {
	var upr UserProjectRole
	err := s.pool.QueryRow(ctx,
		`SELECT upr.user_id, upr.project_id, upr.role_id, pr.name, upr.created_at
		 FROM user_project_roles upr
		 JOIN project_roles pr ON pr.id = upr.role_id
		 WHERE upr.user_id = $1 AND upr.project_id = $2`,
		userID, projectID,
	).Scan(&upr.UserID, &upr.ProjectID, &upr.RoleID, &upr.RoleName, &upr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("rbac: get user role: %w", err)
	}
	return &upr, nil
}

func (s *Service) RemoveUserRole(ctx context.Context, userID, projectID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM user_project_roles WHERE user_id = $1 AND project_id = $2`,
		userID, projectID,
	)
	if err != nil {
		return fmt.Errorf("rbac: remove user role: %w", err)
	}
	return nil
}

func (s *Service) CheckAccess(ctx context.Context, userID, projectID, resource, action string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM user_project_roles upr
			JOIN project_role_permissions prp ON prp.role_id = upr.role_id
			WHERE upr.user_id = $1
			  AND upr.project_id = $2
			  AND (prp.resource = $3 OR prp.resource = '*')
			  AND (prp.action = $4 OR prp.action = '*')
		)`,
		userID, projectID, resource, action,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("rbac: check access: %w", err)
	}
	return exists, nil
}
