# Cosmoria - AGENTS.md

This document defines rules and guidelines for AI agents, automated tools, and contributors interacting with the Cosmoria codebase.

Cosmoria is a backend engine for building multi-tenant applications on PostgreSQL.

---

# 🧠 Core Principles

Any agent working in this repository must respect the following principles:

## 1. Multi-tenancy is mandatory

All data access MUST be scoped by:
- project_id
- tenant_id

Never trust client-provided identity or tenant values.

---

## 2. Backend is the source of truth

The backend is responsible for:
- authentication
- authorization
- tenant isolation
- data validation

Frontend or external clients must never enforce security rules.

---

## 3. Minimal dependencies

Cosmoria prefers:
- standard library usage
- minimal external dependencies
- explicit over abstracted logic

Avoid unnecessary frameworks.

---

## 4. Clear separation of concerns

- `api/` → HTTP layer only
- `internal/` → business logic
- `core/` → system orchestration
- `db/` → database access only

No mixing responsibilities.

---

# 🔐 Security Rules

## Authentication

- JWT or API keys must be validated on every request
- Passwords must be hashed using secure algorithms (e.g., bcrypt or argon2)
- Tokens must have expiration

## API Key Authentication

Cosmoria supports API keys as an alternative to JWT for SaaS user endpoints.

- API keys are attached to a `user_id` — they inherit the user's RBAC role
- Sent via header `X-Api-Key` (separate from `Authorization: Bearer`)
- Format: `ck_<64 hex chars>` (e.g., `ck_a1b2c3d4e5f6...`)
- Stored as SHA-256 hash (never stored in plaintext)
- Plaintext returned only once at creation time
- Created/managed by super_admin via `/api/admin/projects/{pid}/api-keys`
- Cannot bypass RBAC — API keys are subject to `CheckAccess(userID, projectID, resource, action)` like JWT
- Auth middleware checks `Authorization: Bearer` (JWT) first, then falls back to `X-Api-Key`

---

## Authorization

- RBAC must be enforced at request level
- Every action must verify permissions
- Role checks must never be skipped

---

## Data Isolation

Every query MUST include:
- project_id
- tenant_id

Any omission is considered a critical security bug.

---

# 👑 Admin System Rules

Cosmoria has two separate auth systems:

## Platform Admins (`admin_users`)

Administrators of the Cosmoria platform itself (not end-users of built apps).

- `admin_users` table stores: id, email, password_hash, role, created_at
- Roles: `super_admin` (full access), `admin` (limited, assigned per-project)
- Auth via JWT with `ADMIN_JWT_SECRET` (separate from user `JWT_SECRET`)
- API endpoints under `/api/admin/` prefix

## Per-Project Admin Permissions (`admin_project_roles`)

- Granular roles assigned to `admin_users` per-project
- Stored in `admin_project_roles` table (admin_user_id, project_id, role)
- Only `super_admin` can assign/remove roles
- Permissions verified on every admin request by querying DB (not embedded in JWT)

## Bootstrap Flow

- First startup: `POST /api/admin/setup` creates the initial `super_admin` + default project
- Only works once (checks `admin_users` count)
- After setup: `POST /api/admin/login` for admin auth

---

# 🎭 RBAC (Role-Based Access Control) Rules

Cosmoria enforces per-project RBAC for SaaS end-users.

## Role Definitions

- Roles are created per-project by `super_admin` via API
- Each role has a set of `(resource, action)` permissions
- Supported resources: `tenants`, `collections`, `records`, `files`
- Supported actions: `create`, `read`, `update`, `delete`
- Wildcard `*` matches any resource or action

## User-Role Assignment

- Users are assigned to exactly one role per project (`user_project_roles` table)
- Role is NOT embedded in the JWT — queried from DB on every request
- Assignment managed by `super_admin` via `/api/admin/projects/{pid}/users/{uid}/role`

## Enforcement

- RBAC middleware wraps each user-facing route with `RequirePermission(svc, resource, action)`
- Middleware checks `auth.GetAuth(ctx)` for UserID + ProjectID
- Queries `CheckAccess(userID, projectID, resource, action) → bool`
- Supports wildcards: if a role has `(tenants, *)`, it covers all actions on tenants

## Admin routes

RBAC management endpoints are under `/api/admin/` and protected by admin auth (super_admin only).

---

# 📦 Collections System Rules

Collections define dynamic schemas stored in PostgreSQL.

Rules:
- Schema is stored as JSONB
- No direct SQL schema modifications per collection
- All CRUD must go through the collections engine
- Schema mutation is always permitted; existing records are NOT re-validated

---

# 💾 Storage Backend Rules

Cosmoria supports two storage backends: S3-compatible and local filesystem.

## Auto-Detection

- If `S3_ACCESS_KEY` is set AND `Ping()` succeeds → `S3Backend`
- Otherwise → `LocalBackend` (logs a warning on S3 failure)
- No explicit config flag — auto-detection reduces setup steps for local dev

## Backend Interface (`StorageBackend`)

| Method | Purpose |
|--------|---------|
| `Upload(ctx, key, reader, size, contentType)` | Store a file |
| `DownloadURL(ctx, key, expiry)` | Return presigned URL (S3) or metadata path (local) |
| `Delete(ctx, key)` | Remove stored file |
| `Ping(ctx)` | Health check (HEAD bucket for S3, always OK for local) |
| `IsLocal()` | Returns `true` for local, `false` for S3 |

## Local Backend (`LocalBackend`)

- Writes files to `<STORAGE_PATH>/<key>` (default `./data/files`)
- `DownloadURL` returns the key itself; the service constructs the API download URL
- File download endpoint `GET /.../files/{fid}/download` streams bytes (auth + RBAC enforced)
- Path traversal protection via `isPathSafe()` (uses `filepath.Abs` + `filepath.Rel`)

## S3 Backend (`S3Backend`)

- Delegates to `S3Client` (minio-compatible, supports any S3-compatible API)
- `DownloadURL` generates a presigned URL with configurable expiry
- `Ping` performs a HEAD bucket request

## Files Service (`storage.Service`)

- Files are stored with key format: `{projectID}/{tenantID}/{fileID}-{sanitizedFilename}`
- Metadata (id, project_id, tenant_id, filename, s3_key, size, mime_type, uploaded_by) stored in `files` table
- Upload generates a unique suffix + sanitizes filenames
- Delete removes from both storage backend AND database (transactional)

---

# 🔌 WebSocket Realtime Rules

Cosmoria provides realtime event broadcasting via WebSocket backed by PostgreSQL `LISTEN`/`NOTIFY`.

## Architecture

```
Event source → Publisher (pg_notify) → PostgreSQL → Subscriber (LISTEN) → Hub → Client WebSocket
```

| Component | File | Role |
|-----------|------|------|
| `Event` | `events.go` | JSON-serializable payload: id, project_id, tenant_id, resource, action, resource_id, payload, timestamp |
| `Publisher` | `pubsub.go` | Calls `SELECT pg_notify(channel, payload)` in a goroutine |
| `Subscriber` | `pubsub.go` | Dedicated PG connection, dynamically LISTEN/UNLISTEN per-project channels |
| `Hub` | `hub.go` | Manages clients, fan-out events to matching tenants |
| `Client` | `client.go` | Gorilla WebSocket, read/write pumps, ping/pong |

## Channels

- Channel name format: `cosm_{projectId}`
- Dynamic LISTEN/UNLISTEN — channels are added/removed when the first/last client per project connects/disconnects
- `Publisher.Publish()` is called from file upload/delete handlers (future: other resources)

## WebSocket Endpoint

- `GET /api/projects/{pid}/ws?token=<JWT>[&tenant_id=<tid>]`
- Auth via query param `token=` (validates JWT, checks project_id match)
- Optional `tenant_id` — validates user has access via `user_tenants` table; if omitted, receives all project events
- Upgrades to WebSocket, registers client with Hub

## Client Lifecycle

- `writePump`: writes events from `send` channel to WebSocket connection; sends ping every 30s
- `readPump`: reads pong responses; handles incoming "ping" → replies "pong"; closes on error
- On disconnect: unregisters from Hub (removes LISTEN if last client)
- Send buffer: 64 events, drops oldest if full (with warning log)

## Integration

- Handlers publish events after create/delete operations (e.g., files)
- Current resources with events: `files` (create, delete)
- No new dependency for PostgreSQL pub/sub (stdlib pgx only)
- WebSocket dependency: `gorilla/websocket` (approved)

---

# 📋 Audit Logging Rules

Cosmoria records user actions for security and compliance.

## Logger

- `audit.Logger.Log(ctx, projectID, userID, action, resource, resourceID, details, ipAddress)`
- Runs asynchronously in a goroutine — never blocks the request
- Logs errors to slog if insert fails (does NOT fail the request)

## Table (`audit_logs`)

| Column | Type | Purpose |
|--------|------|---------|
| `id` | UUID PK | Auto-generated |
| `project_id` | UUID → projects | Scoping |
| `user_id` | UUID → users | SaaS user who performed the action |
| `action` | TEXT | e.g. "create", "delete", "update", "login" |
| `resource` | TEXT | e.g. "tenants", "records", "files" |
| `resource_id` | UUID (nullable) | ID of the affected resource |
| `details` | JSONB (nullable) | Arbitrary context |
| `ip_address` | TEXT | Client IP from request |
| `created_at` | TIMESTAMPTZ | Auto-set |

## List Endpoint

- `GET /api/admin/projects/{pid}/audit-logs` (admin auth)
- Cursor-based pagination (by `created_at`)
- Max 100 per page, default 50

---

# 📖 OpenAPI / Swagger Rules

Cosmoria generates an OpenAPI spec from Go handler annotations using `swaggo/swag`.

- **Every handler** MUST have swaggo annotations: `@Summary`, `@Description`, `@Tags`, `@Param`, `@Success`, `@Failure`, `@Router`
- **Auth-required routes** MUST include `@Security BearerAuth` or `@Security AdminBearerAuth`
- **Response types** in `{object}` MUST reference the actual Go struct (package.TypeName)
- The spec is served at `/docs/doc.json` and Swagger UI at `/docs/`
- To regenerate after changing annotations: `swag init -g cmd/cosmoria/main.go -o docs/`
- Generated files (`docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`) are committed

---

# 🖥️ CLI Rules

All interaction goes through the `cosmoria` binary.

```
cosmoria serve              Start server (default)
cosmoria dev                Hot reload (watch .go → rebuild → restart)
cosmoria init               Generate .env, docker-compose.yml, Dockerfile
cosmoria migrate new <name> Create migration pair
cosmoria migrate up/down    Run/revert migrations
cosmoria mcp                MCP server (JSON-RPC over stdin/stdout)
```

- The CLI uses only stdlib `flag` and `os.Args` — no cobra/urfave
- `serve` and `migrate` commands load config and connect to the database
- `dev` command compiles and runs a child binary, watching for `.go` file changes
- `init` command creates files in the current directory (not in `cmd/`)
- `mcp` command starts an MCP server (JSON-RPC 2.0 over stdin/stdout)

---

# 📋 Declarative Config Rules (`cosmoria.yaml`)

Cosmoria applies a YAML config file at startup if `cosmoria.yaml` exists in the working directory.

```yaml
project: my-saas
tenants:
  - name: acme-corp
collections:
  - name: posts
    schema:
      fields:
        - { name: title, type: string, required: true }
        - { name: body, type: string }
roles:
  - name: editor
    permissions:
      - { resource: records, action: create }
```

Rules:
- All operations are **idempotent** — resources are matched by name inside the project
- If no `admin_users` exist, `ADMIN_EMAIL` and `ADMIN_PASSWORD` env vars must be set
- The config file is applied **after** migrations, **before** the HTTP server starts
- Does NOT replace the REST API — it's a declarative alternative for initial setup
- Format is YAML only (not JSON, not TOML)
- Config lives in `internal/configfile/`: parser + applier

---

# 🤖 MCP Server Rules

`cosmoria mcp` exposes Cosmoria tools via the Model Context Protocol (JSON-RPC 2.0 over stdin/stdout).

## Transport

- **Stdio**: reads newline-delimited JSON-RPC from stdin, writes responses to stdout
- **No HTTP transport** yet (planned for future)
- Logs go to stderr (never to stdout)

## Handshake (3 steps)

1. Client sends `initialize` → Server responds with capabilities
2. Client sends `notifications/initialized` → Server enters Ready state
3. Server responds to `tools/list` and `tools/call`

## Tools exposed (~21 tools)

| Group | Tools |
|-------|-------|
| Setup | `cosmoria_setup` |
| Projects | `cosmoria_project_create`, `cosmoria_project_list` |
| Tenants | `cosmoria_tenant_create`, `cosmoria_tenant_list`, `cosmoria_tenant_get` |
| Collections | `cosmoria_collection_create`, `cosmoria_collection_list`, `cosmoria_collection_get` |
| RBAC | `cosmoria_role_create`, `cosmoria_role_list`, `cosmoria_role_set_permission`, `cosmoria_role_list_permissions` |
| Records | `cosmoria_record_create`, `cosmoria_record_list`, `cosmoria_record_get`, `cosmoria_record_update`, `cosmoria_record_delete` |
| Users | `cosmoria_user_assign_role` |
| Files | `cosmoria_file_list` |
| Audit | `cosmoria_audit_list` |

## Implementation

- Lives in `internal/mcp/`: types, server (message loop + dispatch), tools (definitions + handlers)
- Each tool handler calls the existing service layer (adminauth, tenant, collections, rbac, records)
- No authentication — the MCP server has direct DB access (local superuser)
- No new dependencies for stdio transport (stdlib only)
- State machine: `New → Initialized → Ready`

# 🎯 TypeScript SDK Rules

The TypeScript SDK lives at `sdk/typescript/`.

- Zero runtime dependencies — uses only `fetch`
- All methods return typed Promises
- Auto-generated `api.ts` from OpenAPI spec (via `scripts/generate-sdk.sh`)
- Hand-written `client.ts` provides the `CosmoriaClient` class
- After adding/modifying API endpoints:
  1. Add swaggo annotations to the handler
  2. Run `swag init -g cmd/cosmoria/main.go -o docs/`
  3. Run `./scripts/generate-sdk.sh`
  4. Update `client.ts` with the new method
- Auth tokens are passed via `client.setToken(token)` between requests
- WebSocket connection: `client.connectWebSocket(projectId, tenantId?)` returns a `WebSocket` for realtime events
- File operations: `client.files.upload(pid, tid, file)`, `client.files.list(pid, tid, cursor?, limit?)`, `client.files.get(pid, tid, fid)`, `client.files.delete(pid, tid, fid)` with typed responses `FileResponse` and `PaginatedFiles`

---

# 🔧 Zero-Config Rules

Cosmoria starts without any environment variables:

| Variable | Default | Behavior |
|----------|---------|----------|
| `DATABASE_URL` | `postgres://localhost:5432/cosmoria?sslmode=disable` | Falls back silently |
| `JWT_SECRET` | Random 32-byte hex | Generated on startup (warns, tokens lost on restart) |
| `ADMIN_JWT_SECRET` | Random 32-byte hex | Generated on startup (warns, tokens lost on restart) |

- Production deployments MUST set both JWT secrets explicitly
- The default DATABASE_URL expects PostgreSQL on localhost:5432
- `cosmoria init` generates a docker-compose.yml with PostgreSQL for this
- Never remove the `slog.Warn` messages — they alert users to the ephemeral secrets

---

# 🔥 Hot Reload Rules

`cosmoria dev` uses `fsnotify` for inotify file watching.

- Watches all `.go` files recursively (excluding `.git` and hidden dirs)
- Debounces changes over 150ms to avoid rapid rebuilds
- On change: kills child process → rebuild → restart
- If build fails: the previous binary is restarted (not left dead)
- The dev binary is compiled to `/tmp/cosmoria-dev`
- Environment variables are inherited from the parent shell

---

# 📦 Dependencies

Approved external dependencies:

| Package | Version | Purpose |
|---------|---------|---------|
| `pgx/v5` | v5.10.0 | PostgreSQL driver |
| `golang-migrate/migrate/v4` | v4.19.1 | Database migrations |
| `golang-jwt/jwt/v5` | v5.3.1 | JWT signing/validation |
| `golang.org/x/crypto` | latest | bcrypt hashing |
| `swaggo/http-swagger` | v1.3.4 | Serve Swagger UI |
| `fsnotify` | v1.10.1 | File watcher for hot reload |
| `gorilla/websocket` | v1.5.3 | WebSocket server for realtime events |
| `yaml.v3` | v3.0.1 | YAML config file parsing |

Dev-only tools:
| Tool | Purpose |
|------|---------|
| `swaggo/swag` CLI | Generate OpenAPI spec from annotations |
| `openapi-typescript` (npm) v5 | Generate TS types from OpenAPI spec (v5 required for Swagger 2.0 compat) |

Always consult the approved list before adding new dependencies.

---

# ⚡ API Design Rules

- All endpoints must be deterministic
- Pagination is mandatory for list endpoints
- No unbounded queries allowed
- Error responses must be consistent

---

# 🧩 Code Style Rules

- Prefer explicit code over magic abstraction
- Keep functions small and focused
- Avoid hidden side effects
- Use clear naming conventions

---

# 🧪 Testing Expectations

Agents should ensure:

- tenant isolation is always tested
- auth cannot be bypassed
- API responses are consistent
- edge cases are handled safely

---

# 🚫 Forbidden Actions

Agents MUST NOT:

- bypass authentication checks
- ignore tenant or project scoping
- introduce global state without reason
- add unnecessary dependencies
- mix API layer with business logic

---

# 🏗️ Architecture Awareness

Cosmoria is not a traditional web app.

It is:

> A backend infrastructure engine for multi-tenant SaaS applications.

All contributions must respect this architectural vision.

---

# 🎯 Goal of the Project

To build a production-grade, open-source backend engine that enables developers to create SaaS applications quickly with:

- dynamic APIs
- multi-tenancy
- built-in storage integration
- authentication and RBAC