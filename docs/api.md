# Cosmoria — API Reference

All endpoints are served under the configured port (default `:8080`).  
OpenAPI spec available at `/openapi.json` and Swagger UI at `/docs/`.

---

## Public Routes (no auth required)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/openapi.json` | OpenAPI spec (redirects to `/docs/doc.json`) |
| GET | `/docs/` | Swagger UI |
| GET | `/docs/{asset}` | Swagger UI assets (CSS, JS, doc.json) |
| POST | `/api/auth/signup` | Register a new SaaS user |
| POST | `/api/auth/login` | Authenticate and receive JWT |
| POST | `/api/admin/setup` | Bootstrap first super_admin + project (once only) |
| POST | `/api/admin/login` | Admin login |

---

## Admin Routes (require `AdminBearerAuth` JWT)

### Projects

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/admin/projects` | Create a new project |
| GET | `/api/admin/projects` | List accessible projects |
| GET | `/api/admin/projects/{pid}` | Get a single project |
| PUT | `/api/admin/projects/{pid}` | Update project name or JWT expiry |
| DELETE | `/api/admin/projects/{pid}` | Delete project and all data (super_admin only) |

### API Keys (super_admin only)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/admin/projects/{pid}/api-keys` | Create API key for a user |
| GET | `/api/admin/projects/{pid}/api-keys` | List API keys |
| DELETE | `/api/admin/projects/{pid}/api-keys/{kid}` | Revoke an API key |

### Audit Logs

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/admin/projects/{pid}/audit-logs` | List audit entries (cursor pagination) |

### Admin Roles (super_admin only)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/admin/projects/{pid}/admin-roles` | Assign an admin to a project |
| GET | `/api/admin/projects/{pid}/admin-roles` | List admin roles for a project |
| DELETE | `/api/admin/projects/{pid}/admin-roles/{aid}` | Remove an admin's project access |

### RBAC Roles (super_admin only)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/admin/projects/{pid}/roles` | Create a new RBAC role |
| GET | `/api/admin/projects/{pid}/roles` | List all RBAC roles for a project |
| DELETE | `/api/admin/projects/{pid}/roles/{rid}` | Delete an RBAC role |
| POST | `/api/admin/projects/{pid}/roles/{rid}/permissions` | Add a permission to a role |
| DELETE | `/api/admin/projects/{pid}/roles/{rid}/permissions` | Remove a permission from a role |
| GET | `/api/admin/projects/{pid}/roles/{rid}/permissions` | List permissions for a role |
| POST | `/api/admin/projects/{pid}/users/{uid}/role` | Assign an RBAC role to a SaaS user |
| GET | `/api/admin/projects/{pid}/users/{uid}/role` | Get a user's assigned RBAC role |
| DELETE | `/api/admin/projects/{pid}/users/{uid}/role` | Remove a user's RBAC role |

### Collections (super_admin only)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/admin/projects/{pid}/collections` | Create a collection with schema |
| GET | `/api/admin/projects/{pid}/collections` | List all collections |
| GET | `/api/admin/projects/{pid}/collections/{cid}` | Get a collection definition |
| PUT | `/api/admin/projects/{pid}/collections/{cid}` | Update a collection's schema |
| DELETE | `/api/admin/projects/{pid}/collections/{cid}` | Delete a collection and its records |

---

## User Routes (require `BearerAuth` JWT)

### Tenants

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/projects/{pid}/tenants` | Create a tenant |
| GET | `/api/projects/{pid}/tenants` | List tenants |
| GET | `/api/projects/{pid}/tenants/{tid}` | Get a tenant |
| DELETE | `/api/projects/{pid}/tenants/{tid}` | Delete a tenant |
| POST | `/api/projects/{pid}/tenants/{tid}/users` | Assign a user to a tenant |
| DELETE | `/api/projects/{pid}/tenants/{tid}/users/{uid}` | Remove a user from a tenant |

### Records

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/projects/{pid}/tenants/{tid}/collections/{cid}/records` | Create a record |
| GET | `/api/projects/{pid}/tenants/{tid}/collections/{cid}/records` | List records (cursor pagination) |
| GET | `/api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}` | Get a record |
| PUT | `/api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}` | Update a record (full replacement) |
| DELETE | `/api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}` | Delete a record |

**Query parameters for List records:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `cursor` | string | — | Record ID for cursor-based pagination |
| `limit` | int | 50 | Page size (max 100) |

Response format:
```json
{
  "data": [ { "id": "...", "data": { ... }, "created_at": "..." } ],
  "next_cursor": "uuid"  // omitted if no more pages
}
```

### User Profile

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/auth/me` | Get authenticated user's profile |
| PUT | `/api/auth/me` | Update email and/or password (requires `current_password`) |

### Files (RBAC: `files` resource)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/projects/{pid}/tenants/{tid}/files` | Upload a file (multipart form) |
| GET | `/api/projects/{pid}/tenants/{tid}/files` | List files (cursor pagination) |
| GET | `/api/projects/{pid}/tenants/{tid}/files/{fid}` | Get file metadata + download URL |
| DELETE | `/api/projects/{pid}/tenants/{tid}/files/{fid}` | Delete a file |
| GET | `/api/projects/{pid}/tenants/{tid}/files/{fid}/download` | Stream file content (local backend only) |

**Query parameters for List files:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `cursor` | string | — | CreatedAt cursor for pagination |
| `limit` | int | 50 | Page size (max 100) |

### Realtime

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/projects/{pid}/ws` | WebSocket (query param `token=JWT`, optional `tenant_id`) |

The WebSocket endpoint expects authentication via `?token=` query parameter (JWT). If `?tenant_id=` is provided, only events for that tenant are received; otherwise all project events are broadcast. Supports `{"type": "ping"}` → `{"type": "pong"}` keep-alive.

---

## Authentication Notes

- User auth: `Authorization: Bearer <JWT>` or `X-Api-Key: ck_<64hex>` header
- Admin auth: `Authorization: Bearer <admin JWT>` only (no API key support)
- API keys are attached to a user and inherit that user's RBAC role
