export class CosmoriaClient {
  private baseUrl: string;
  private token?: string;

  constructor(baseUrl: string, token?: string) {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.token = token;
  }

  setToken(token: string) {
    this.token = token;
  }

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };

    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    const res = await fetch(`${this.baseUrl}${path}`, {
      method,
      headers,
      body: body ? JSON.stringify(body) : undefined,
    });

    if (res.status === 204) {
      return undefined as T;
    }

    const data = await res.json();

    if (!res.ok) {
      throw new CosmoriaError(res.status, (data as ApiError).error ?? "unknown_error", data);
    }

    return data as T;
  }

  /** Health check */
  health() {
    return this.request<{ status: string }>("GET", "/health");
  }

  /** WebSocket realtime connection */
  connectWebSocket(
    projectId: string,
    token: string,
    tenantId?: string,
  ): WebSocket {
    const params = new URLSearchParams({ token });
    if (tenantId) params.set("tenant_id", tenantId);
    return new WebSocket(
      `${this.baseUrl.replace(/^http/, "ws")}/api/projects/${projectId}/ws?${params}`,
    );
  }

  /** Auth */
  auth = {
    signup: (email: string, password: string, project_id: string) =>
      this.request<AuthResult>("POST", "/api/auth/signup", { email, password, project_id }),

    login: (email: string, password: string, project_id: string) =>
      this.request<AuthResult>("POST", "/api/auth/login", { email, password, project_id }),
  };

  /** Admin */
  admin = {
    setup: (email: string, password: string) =>
      this.request<{ token: string; admin: AdminUser; project: Project }>(
        "POST", "/api/admin/setup", { email, password },
      ),

    login: (email: string, password: string) =>
      this.request<AdminAuthResult>("POST", "/api/admin/login", { email, password }),

    createProject: (name: string) =>
      this.request<Project>("POST", "/api/admin/projects", { name }),

    listProjects: () =>
      this.request<ProjectWithRole[]>("GET", "/api/admin/projects"),

    assignAdminRole: (projectId: string, adminUserId: string, role: string) =>
      this.request<void>("POST", `/api/admin/projects/${projectId}/admin-roles`, {
        admin_user_id: adminUserId, role,
      }),

    listAdminRoles: (projectId: string) =>
      this.request<AdminProjectRole[]>("GET", `/api/admin/projects/${projectId}/admin-roles`),

    removeAdminRole: (projectId: string, adminUserId: string) =>
      this.request<void>("DELETE", `/api/admin/projects/${projectId}/admin-roles/${adminUserId}`),
  };

  /** Tenants */
  tenants = {
    create: (projectId: string, name: string) =>
      this.request<Tenant>("POST", `/api/projects/${projectId}/tenants`, { name }),

    list: (projectId: string) =>
      this.request<Tenant[]>("GET", `/api/projects/${projectId}/tenants`),

    get: (projectId: string, tenantId: string) =>
      this.request<Tenant>("GET", `/api/projects/${projectId}/tenants/${tenantId}`),

    delete: (projectId: string, tenantId: string) =>
      this.request<void>("DELETE", `/api/projects/${projectId}/tenants/${tenantId}`),

    assignUser: (projectId: string, tenantId: string, userId: string) =>
      this.request<void>("POST", `/api/projects/${projectId}/tenants/${tenantId}/users`, {
        user_id: userId,
      }),

    removeUser: (projectId: string, tenantId: string, userId: string) =>
      this.request<void>(
        "DELETE", `/api/projects/${projectId}/tenants/${tenantId}/users/${userId}`,
      ),
  };

  /** RBAC */
  rbac = {
    createRole: (projectId: string, name: string) =>
      this.request<RbacRole>("POST", `/api/admin/projects/${projectId}/roles`, { name }),

    listRoles: (projectId: string) =>
      this.request<RbacRoleWithPermissions[]>(
        "GET", `/api/admin/projects/${projectId}/roles`,
      ),

    deleteRole: (projectId: string, roleId: string) =>
      this.request<void>("DELETE", `/api/admin/projects/${projectId}/roles/${roleId}`),

    setPermission: (roleId: string, resource: string, action: string) =>
      this.request<RbacPermission>(
        "POST", `/api/admin/projects/${roleId}/permissions`,
        { resource, action },
      ),

    removePermission: (roleId: string, resource: string, action: string) =>
      this.request<void>(
        "DELETE", `/api/admin/projects/${roleId}/permissions`,
        { resource, action },
      ),

    listPermissions: (roleId: string) =>
      this.request<RbacPermission[]>("GET", `/api/admin/projects/${roleId}/permissions`),

    assignUserRole: (projectId: string, userId: string, roleId: string) =>
      this.request<UserProjectRole>(
        "POST", `/api/admin/projects/${projectId}/users/${userId}/role`,
        { role_id: roleId },
      ),

    getUserRole: (projectId: string, userId: string) =>
      this.request<UserProjectRole>(
        "GET", `/api/admin/projects/${projectId}/users/${userId}/role`,
      ),

    removeUserRole: (projectId: string, userId: string) =>
      this.request<void>(
        "DELETE", `/api/admin/projects/${projectId}/users/${userId}/role`,
      ),
  };

  /** Collections */
  collections = {
    create: (projectId: string, name: string, schema: CollectionSchema) =>
      this.request<Collection>(
        "POST", `/api/admin/projects/${projectId}/collections`,
        { name, schema },
      ),

    list: (projectId: string) =>
      this.request<Collection[]>("GET", `/api/admin/projects/${projectId}/collections`),

    get: (projectId: string, collectionId: string) =>
      this.request<Collection>(
        "GET", `/api/admin/projects/${projectId}/collections/${collectionId}`,
      ),

    updateSchema: (projectId: string, collectionId: string, schema: CollectionSchema) =>
      this.request<Collection>(
        "PUT", `/api/admin/projects/${projectId}/collections/${collectionId}`,
        { schema },
      ),

    delete: (projectId: string, collectionId: string) =>
      this.request<void>(
        "DELETE", `/api/admin/projects/${projectId}/collections/${collectionId}`,
      ),
  };

  /** Records */
  records = {
    create: (
      projectId: string, tenantId: string, collectionId: string, data: RecordData,
    ) =>
      this.request<Record>(
        "POST",
        `/api/projects/${projectId}/tenants/${tenantId}/collections/${collectionId}/records`,
        { data },
      ),

    list: (
      projectId: string, tenantId: string, collectionId: string,
      cursor?: string, limit?: number,
    ) =>
      this.request<PaginatedRecords>(
        "GET",
        `/api/projects/${projectId}/tenants/${tenantId}/collections/${collectionId}/records`
        + `?${new URLSearchParams({ ...(cursor ? { cursor } : {}), ...(limit ? { limit: String(limit) } : {}) })}`,
      ),

    get: (projectId: string, tenantId: string, collectionId: string, recordId: string) =>
      this.request<Record>(
        "GET",
        `/api/projects/${projectId}/tenants/${tenantId}/collections/${collectionId}/records/${recordId}`,
      ),

    update: (
      projectId: string, tenantId: string, collectionId: string,
      recordId: string, data: RecordData,
    ) =>
      this.request<Record>(
        "PUT",
        `/api/projects/${projectId}/tenants/${tenantId}/collections/${collectionId}/records/${recordId}`,
        { data },
      ),

    delete: (
      projectId: string, tenantId: string, collectionId: string, recordId: string,
    ) =>
      this.request<void>(
        "DELETE",
        `/api/projects/${projectId}/tenants/${tenantId}/collections/${collectionId}/records/${recordId}`,
      ),
  };

  /** Files */
  files = {
    upload: (projectId: string, tenantId: string, file: File | Blob, filename?: string) =>
      this.uploadFile<FileResponse>(
        `/api/projects/${projectId}/tenants/${tenantId}/files`,
        file, filename,
      ),

    list: (
      projectId: string, tenantId: string,
      cursor?: string, limit?: number,
    ) =>
      this.request<PaginatedFiles>(
        "GET",
        `/api/projects/${projectId}/tenants/${tenantId}/files`
        + `?${new URLSearchParams({ ...(cursor ? { cursor } : {}), ...(limit ? { limit: String(limit) } : {}) })}`,
      ),

    get: (projectId: string, tenantId: string, fileId: string) =>
      this.request<FileResponse>(
        "GET",
        `/api/projects/${projectId}/tenants/${tenantId}/files/${fileId}`,
      ),

    delete: (projectId: string, tenantId: string, fileId: string) =>
      this.request<void>(
        "DELETE",
        `/api/projects/${projectId}/tenants/${tenantId}/files/${fileId}`,
      ),
  };

  private async uploadFile<T>(
    path: string,
    file: File | Blob,
    filename?: string,
  ): Promise<T> {
    const fd = new FormData();
    fd.append("file", file, filename);
    const headers: Record<string, string> = {};
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }
    const res = await fetch(`${this.baseUrl}${path}`, {
      method: "POST",
      headers,
      body: fd,
    });
    const data = await res.json();
    if (!res.ok) {
      throw new CosmoriaError(res.status, (data as ApiError).error ?? "unknown_error", data);
    }
    return data as T;
  }
}

// ---- Types ----

export interface ApiError {
  error: string;
}

export interface AuthResult {
  token: string;
  user: {
    id: string;
    email: string;
    project_id: string;
  };
}

export interface AdminUser {
  id: string;
  email: string;
  role: string;
  created_at: string;
}

export interface Project {
  id: string;
  name: string;
  admin_owner_id: string;
  jwt_expiry?: number;
  created_at: string;
}

export interface ProjectWithRole extends Project {
  role: string;
}

export interface AdminProjectRole {
  admin_user_id: string;
  project_id: string;
  role: string;
  created_at: string;
}

export interface AdminAuthResult {
  token: string;
  admin: AdminUser;
}

export interface Tenant {
  id: string;
  name: string;
  project_id: string;
  created_at: string;
}

export interface RbacRole {
  id: string;
  project_id: string;
  name: string;
  created_at: string;
}

export interface RbacPermission {
  id: string;
  role_id: string;
  resource: string;
  action: string;
}

export interface RbacRoleWithPermissions extends RbacRole {
  permissions: RbacPermission[];
}

export interface UserProjectRole {
  user_id: string;
  project_id: string;
  role_id: string;
  created_at: string;
}

export interface CollectionField {
  name: string;
  type: "string" | "number" | "boolean";
  required: boolean;
}

export interface CollectionSchema {
  fields: CollectionField[];
}

export interface Collection {
  id: string;
  project_id: string;
  name: string;
  schema: CollectionSchema;
  created_at: string;
}

export type RecordData = Record<string, unknown>;

export interface Record {
  id: string;
  data: RecordData;
  created_at: string;
}

export interface PaginatedRecords {
  data: Record[];
  next_cursor?: string;
}

export interface FileResponse {
  id: string;
  project_id: string;
  tenant_id: string;
  filename: string;
  mime_type: string;
  size: number;
  presigned_url?: string;
  uploaded_by: string;
  created_at: string;
}

export interface PaginatedFiles {
  files: FileResponse[];
  next_cursor?: string;
}

export class CosmoriaError extends Error {
  status: number;
  body: unknown;

  constructor(status: number, message: string, body: unknown) {
    super(message);
    this.name = "CosmoriaError";
    this.status = status;
    this.body = body;
  }
}
