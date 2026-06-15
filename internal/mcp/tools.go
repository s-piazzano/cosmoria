package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/s-piazzano/cosmoria/internal/collections"
)

func tools() []Tool {
	return []Tool{
		{
			Name:        "cosmoria_setup",
			Description: "Create the initial super_admin account and default project. Only works once.",
			InputSchema: objSchema(
				jsonProp{"email", "string", "Admin email address", true},
				jsonProp{"password", "string", "Admin password", true},
				jsonProp{"project_name", "string", "Name of the initial project", true},
			),
			Handler: handleSetup,
		},
		{
			Name:        "cosmoria_project_create",
			Description: "Create a new project. Requires an admin user to exist (run cosmoria_setup first).",
			InputSchema: objSchema(
				jsonProp{"admin_id", "string", "Admin user UUID", true},
				jsonProp{"name", "string", "Project name", true},
			),
			Handler: handleProjectCreate,
		},
		{
			Name:        "cosmoria_project_list",
			Description: "List all projects in the system. Requires an admin ID.",
			InputSchema: objSchema(
				jsonProp{"admin_id", "string", "Admin user UUID", true},
			),
			Handler:     handleProjectList,
		},
		{
			Name:        "cosmoria_tenant_create",
			Description: "Create a new tenant in a project.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"name", "string", "Tenant name (e.g. 'acme-corp')", true},
			),
			Handler: handleTenantCreate,
		},
		{
			Name:        "cosmoria_tenant_list",
			Description: "List all tenants in a project.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
			),
			Handler: handleTenantList,
		},
		{
			Name:        "cosmoria_tenant_get",
			Description: "Get a single tenant by ID.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"tenant_id", "string", "Tenant UUID", true},
			),
			Handler: handleTenantGet,
		},
		{
			Name:        "cosmoria_collection_create",
			Description: "Create a collection with a dynamic schema in a project.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"name", "string", "Collection name (e.g. 'posts')", true},
				jsonProp{"schema", "object", "Schema definition with fields array. Each field: {name, type: string|number|boolean, required: bool}", true},
			),
			Handler: handleCollectionCreate,
		},
		{
			Name:        "cosmoria_collection_list",
			Description: "List all collections in a project.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
			),
			Handler: handleCollectionList,
		},
		{
			Name:        "cosmoria_collection_get",
			Description: "Get a single collection with its schema by ID.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"collection_id", "string", "Collection UUID", true},
			),
			Handler: handleCollectionGet,
		},
		{
			Name:        "cosmoria_role_create",
			Description: "Create a new RBAC role in a project.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"name", "string", "Role name (e.g. 'editor', 'viewer')", true},
			),
			Handler: handleRoleCreate,
		},
		{
			Name:        "cosmoria_role_list",
			Description: "List all RBAC roles with their permissions in a project.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
			),
			Handler: handleRoleList,
		},
		{
			Name:        "cosmoria_role_set_permission",
			Description: "Add or update a permission on a role. Supports wildcard '*' for resource or action.",
			InputSchema: objSchema(
				jsonProp{"role_id", "string", "Role UUID", true},
				jsonProp{"resource", "string", "Resource name (tenants|collections|records|files|*)", true},
				jsonProp{"action", "string", "Action name (create|read|update|delete|*)", true},
			),
			Handler: handleRoleSetPermission,
		},
		{
			Name:        "cosmoria_role_list_permissions",
			Description: "List all permissions on a role.",
			InputSchema: objSchema(
				jsonProp{"role_id", "string", "Role UUID", true},
			),
			Handler: handleRoleListPermissions,
		},
		{
			Name:        "cosmoria_record_create",
			Description: "Create a new record in a collection. The data must match the collection schema.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"tenant_id", "string", "Tenant UUID", true},
				jsonProp{"collection_id", "string", "Collection UUID", true},
				jsonProp{"data", "object", "Record data as JSON object matching the collection schema", true},
			),
			Handler: handleRecordCreate,
		},
		{
			Name:        "cosmoria_record_list",
			Description: "List records in a collection with cursor-based pagination.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"tenant_id", "string", "Tenant UUID", true},
				jsonProp{"collection_id", "string", "Collection UUID", true},
				jsonProp{"limit", "number", "Max results (1-100, default 50)", false},
				jsonProp{"cursor", "string", "Pagination cursor from previous response", false},
			),
			Handler: handleRecordList,
		},
		{
			Name:        "cosmoria_record_get",
			Description: "Get a single record by ID.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"tenant_id", "string", "Tenant UUID", true},
				jsonProp{"collection_id", "string", "Collection UUID", true},
				jsonProp{"record_id", "string", "Record UUID", true},
			),
			Handler: handleRecordGet,
		},
		{
			Name:        "cosmoria_record_update",
			Description: "Replace the data of an existing record. Must match the collection schema.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"tenant_id", "string", "Tenant UUID", true},
				jsonProp{"collection_id", "string", "Collection UUID", true},
				jsonProp{"record_id", "string", "Record UUID", true},
				jsonProp{"data", "object", "New record data as JSON object", true},
			),
			Handler: handleRecordUpdate,
		},
		{
			Name:        "cosmoria_record_delete",
			Description: "Delete a record by ID.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"tenant_id", "string", "Tenant UUID", true},
				jsonProp{"collection_id", "string", "Collection UUID", true},
				jsonProp{"record_id", "string", "Record UUID", true},
			),
			Handler: handleRecordDelete,
		},
		{
			Name:        "cosmoria_user_assign_role",
			Description: "Assign an RBAC role to a user within a project.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"user_id", "string", "User UUID", true},
				jsonProp{"role_id", "string", "Role UUID", true},
			),
			Handler: handleUserAssignRole,
		},
		{
			Name:        "cosmoria_file_list",
			Description: "List files for a tenant with cursor-based pagination.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"tenant_id", "string", "Tenant UUID", true},
				jsonProp{"limit", "number", "Max results (1-100, default 50)", false},
				jsonProp{"cursor", "string", "Pagination cursor from previous response", false},
			),
			Handler: handleFileList,
		},
		{
			Name:        "cosmoria_audit_list",
			Description: "List audit logs for a project with cursor-based pagination.",
			InputSchema: objSchema(
				jsonProp{"project_id", "string", "Project UUID", true},
				jsonProp{"limit", "number", "Max results (1-100, default 50)", false},
				jsonProp{"cursor", "string", "Pagination cursor from previous response", false},
			),
			Handler: handleAuditList,
		},
	}
}

func handleSetup(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		ProjectName string `json:"project_name"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	result, project, err := svc.AdminAuth.Setup(ctx, p.Email, p.Password, p.ProjectName)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(map[string]any{
		"admin_id": result.Admin.ID,
		"project_id": project.ID,
		"token":   result.Token,
	})
	return string(out), nil
}

func handleProjectCreate(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		AdminID string `json:"admin_id"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	project, err := svc.AdminAuth.CreateProject(ctx, p.AdminID, p.Name)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(map[string]string{
		"id":   project.ID,
		"name": project.Name,
	})
	return string(out), nil
}

func handleProjectList(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		AdminID string `json:"admin_id"`
		Role    string `json:"role"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}
	if p.AdminID == "" {
		return "", fmt.Errorf("admin_id is required")
	}
	if p.Role == "" {
		p.Role = "super_admin"
	}
	projects, err := svc.AdminAuth.ListAccessibleProjects(ctx, p.AdminID, p.Role)
	if err != nil {
		return "", err
	}
	out, _ := json.Marshal(projects)
	return string(out), nil
}

func handleTenantCreate(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	t, err := svc.Tenant.CreateTenant(ctx, p.ProjectID, p.Name)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(t)
	return string(out), nil
}

func handleTenantList(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	tenants, err := svc.Tenant.ListTenants(ctx, p.ProjectID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(tenants)
	return string(out), nil
}

func handleTenantGet(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
		TenantID  string `json:"tenant_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	t, err := svc.Tenant.GetTenant(ctx, p.TenantID, p.ProjectID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(t)
	return string(out), nil
}

func handleCollectionCreate(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string          `json:"project_id"`
		Name      string          `json:"name"`
		SchemaRaw json.RawMessage `json:"schema"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	var sch collections.Schema
	if err := json.Unmarshal(p.SchemaRaw, &sch); err != nil {
		return "", fmt.Errorf("invalid schema: %w", err)
	}

	c, err := svc.Collections.CreateCollection(ctx, p.ProjectID, p.Name, sch)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(c)
	return string(out), nil
}

func handleCollectionList(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	cols, err := svc.Collections.ListCollections(ctx, p.ProjectID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(cols)
	return string(out), nil
}

func handleCollectionGet(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID    string `json:"project_id"`
		CollectionID string `json:"collection_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	c, err := svc.Collections.GetCollection(ctx, p.CollectionID, p.ProjectID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(c)
	return string(out), nil
}

func handleRoleCreate(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
		Name      string `json:"name"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	r, err := svc.RBAC.CreateRole(ctx, p.ProjectID, p.Name)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(r)
	return string(out), nil
}

func handleRoleList(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	roles, err := svc.RBAC.ListRoles(ctx, p.ProjectID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(roles)
	return string(out), nil
}

func handleRoleSetPermission(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		RoleID   string `json:"role_id"`
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	perm, err := svc.RBAC.SetPermission(ctx, p.RoleID, p.Resource, p.Action)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(perm)
	return string(out), nil
}

func handleRoleListPermissions(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		RoleID string `json:"role_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	perms, err := svc.RBAC.ListPermissions(ctx, p.RoleID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(perms)
	return string(out), nil
}

func handleRecordCreate(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID    string         `json:"project_id"`
		TenantID     string         `json:"tenant_id"`
		CollectionID string         `json:"collection_id"`
		Data         map[string]any `json:"data"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	r, err := svc.Records.CreateRecord(ctx, p.ProjectID, p.TenantID, p.CollectionID, p.Data)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(r)
	return string(out), nil
}

func handleRecordList(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID    string `json:"project_id"`
		TenantID     string `json:"tenant_id"`
		CollectionID string `json:"collection_id"`
		Limit        int    `json:"limit"`
		Cursor       string `json:"cursor"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	if p.Limit <= 0 {
		p.Limit = 50
	}

	recs, nextCursor, err := svc.Records.ListRecords(ctx, p.ProjectID, p.TenantID, p.CollectionID, p.Cursor, p.Limit)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(map[string]any{
		"records":     recs,
		"next_cursor": nextCursor,
	})
	return string(out), nil
}

func handleRecordGet(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID    string `json:"project_id"`
		TenantID     string `json:"tenant_id"`
		CollectionID string `json:"collection_id"`
		RecordID     string `json:"record_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	r, err := svc.Records.GetRecord(ctx, p.RecordID, p.ProjectID, p.TenantID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(r)
	return string(out), nil
}

func handleRecordUpdate(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID    string         `json:"project_id"`
		TenantID     string         `json:"tenant_id"`
		CollectionID string         `json:"collection_id"`
		RecordID     string         `json:"record_id"`
		Data         map[string]any `json:"data"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	r, err := svc.Records.UpdateRecord(ctx, p.RecordID, p.ProjectID, p.TenantID, p.Data)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(r)
	return string(out), nil
}

func handleRecordDelete(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID    string `json:"project_id"`
		TenantID     string `json:"tenant_id"`
		CollectionID string `json:"collection_id"`
		RecordID     string `json:"record_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	if err := svc.Records.DeleteRecord(ctx, p.RecordID, p.ProjectID, p.TenantID); err != nil {
		return "", err
	}

	return fmt.Sprintf("Deleted record %s", p.RecordID), nil
}

func handleUserAssignRole(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
		UserID    string `json:"user_id"`
		RoleID    string `json:"role_id"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	upr, err := svc.RBAC.AssignUserRole(ctx, p.UserID, p.ProjectID, p.RoleID)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(upr)
	return string(out), nil
}

func handleFileList(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
		TenantID  string `json:"tenant_id"`
		Limit     int    `json:"limit"`
		Cursor    string `json:"cursor"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	if p.Limit <= 0 {
		p.Limit = 50
	}

	files, nextCursor, err := svc.Storage.List(ctx, p.ProjectID, p.TenantID, p.Cursor, p.Limit)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(map[string]any{
		"files":       files,
		"next_cursor": nextCursor,
	})
	return string(out), nil
}

func handleAuditList(ctx context.Context, args json.RawMessage, svc *Services) (string, error) {
	var p struct {
		ProjectID string `json:"project_id"`
		Limit     int    `json:"limit"`
		Cursor    string `json:"cursor"`
	}
	if err := json.Unmarshal(args, &p); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	if p.Limit <= 0 {
		p.Limit = 50
	}

	entries, nextCursor, err := svc.Audit.List(ctx, p.ProjectID, p.Cursor, p.Limit)
	if err != nil {
		return "", err
	}

	out, _ := json.Marshal(map[string]any{
		"audit_logs":  entries,
		"next_cursor": nextCursor,
	})
	return string(out), nil
}
