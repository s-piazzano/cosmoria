package rbac

import "time"

type Role struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Permission struct {
	ID       string `json:"id"`
	RoleID   string `json:"role_id"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

type UserProjectRole struct {
	UserID    string    `json:"user_id"`
	ProjectID string    `json:"project_id"`
	RoleID    string    `json:"role_id"`
	RoleName  string    `json:"role_name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type RoleWithPermissions struct {
	Role
	Permissions []Permission `json:"permissions"`
}
