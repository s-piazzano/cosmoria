# Cosmoria Architecture

## Overview

Cosmoria is a backend engine for building multi-tenant SaaS applications on PostgreSQL. It provides dynamic APIs, multi-tenancy, authentication and RBAC out of the box.

---

## Directory Structure

```
cmd/cosmoria/main.go          # Entry point: wires config → pool → migrations → services → handlers → middleware
internal/
  adminauth/                  # Platform admin auth (admin_users + admin_project_roles)
    hash.go                   # bcrypt hash/check
    jwt.go                    # Admin JWT generation/validation (ADMIN_JWT_SECRET)
    context.go                # WithAdminAuth / GetAdminAuth (context)
    service.go                # AdminService: Setup, Login, CreateProject, role management
  api/
    router.go                 # Router wrapping http.ServeMux
    handlers/
      health.go               # GET /health
      auth.go                 # AuthHandler: signup, login (SaaS users)
      tenants.go              # TenantHandler: CRUD + user assignment
      admin.go                # AdminHandler: setup, login, project CRUD, roles
    middleware/
      logging.go              # Request logging (method, path, status, duration)
      auth.go                 # User JWT validation, skips /api/admin/
      admin.go                # Admin JWT validation (ADMIN_JWT_SECRET)
      tenant.go               # X-Tenant-ID extraction + access validation
      chain.go                # Chain(handler, mws...) compositor
  auth/                       # SaaS user auth
    hash.go                   # bcrypt hash/check
    jwt.go                    # User JWT (JWT_SECRET)
    context.go                # WithAuth / GetAuth
    service.go                # AuthService: signup, login
  core/
    config.go                 # Config struct (env-based)
    app.go                    # App {Config, Pool, Handler} — Run / Shutdown
  db/
    postgres.go               # NewPool: pgxpool with ping fail-fast
    migrate.go                # Migrate: golang-migrate file source + postgres
  tenant/
    context.go                # WithTenant / GetTenant
    service.go                # TenantService: CRUD + HasAccess, AssignUser, RemoveUser
db/migrations/                # Timestamped .up.sql / .down.sql pairs
docs/architecture.md          # This file
AGENTS.md                     # Rules and guidelines for AI agents
```

---

## Authentication: Two Separate Systems

### 1. Platform Admin Auth (`admin_users`)

Who can admin the Cosmoria platform.

| Concept | Detail |
|---------|--------|
| Table | `admin_users` (id, email, password_hash, role, created_at) |
| Roles | `super_admin` (full), `admin` (limited, per-project) |
| Auth method | JWT with `ADMIN_JWT_SECRET` (env var) |
| Endpoints | Under `/api/admin/` prefix |
| Per-project permissions | `admin_project_roles` table (admin_user_id, project_id, role) |
| Permission model | Not embedded in JWT — queried from DB on every admin request |
| Expiry | `ADMIN_JWT_EXPIRY` env var (default 3600s) |

### 2. SaaS User Auth (`users`)

End-users of applications built on Cosmoria.

| Concept | Detail |
|---------|--------|
| Table | `users` (id, email, password_hash, role, created_at) |
| Roles | `viewer`, `editor`, `admin` |
| Auth method | JWT with `JWT_SECRET` (env var) |
| Endpoints | Under `/api/auth/` prefix |
| Scope | Always scoped to a `project_id` (passed at signup) |
| Expiry | Per-project `jwt_expiry` column in `projects` table, fallback to `JWT_EXPIRY` (default 86400) |

---

## Data Model

### Core Entities

```
admin_users (platform admins)
  id UUID PK
  email TEXT UNIQUE
  password_hash TEXT
  role TEXT (super_admin | admin)
  created_at TIMESTAMPTZ

projects
  id UUID PK
  name TEXT
  admin_owner_id UUID → admin_users(id)
  jwt_expiry BIGINT (nullable, per-project JWT expiry)
  created_at TIMESTAMPTZ

admin_project_roles
  admin_user_id UUID → admin_users(id)
  project_id UUID → projects(id)
  role TEXT
  created_at TIMESTAMPTZ
  PK (admin_user_id, project_id)

users (SaaS end-users)
  id UUID PK
  email TEXT UNIQUE
  password_hash TEXT
  role TEXT (viewer | editor | admin)
  created_at TIMESTAMPTZ

api_keys
  id UUID PK
  project_id UUID → projects(id)
  key_hash TEXT
  name TEXT
  created_at TIMESTAMPTZ

tenants
  id UUID PK
  project_id UUID → projects(id)
  name TEXT
  created_at TIMESTAMPTZ

user_tenants (user tenant access)
  user_id UUID → users(id)
  tenant_id UUID → tenants(id)
  project_id UUID → projects(id)
  created_at TIMESTAMPTZ
  PK (user_id, tenant_id)

collections
  id UUID PK
  project_id UUID → projects(id)
  name TEXT
  schema JSONB
  created_at TIMESTAMPTZ

records
  id UUID PK
  collection_id UUID → collections(id)
  project_id UUID → projects(id)
  data JSONB
  created_at TIMESTAMPTZ

files
  id UUID PK
  project_id UUID → projects(id)
  filename TEXT
  storage_path TEXT
  mime_type TEXT
  size BIGINT
  created_at TIMESTAMPTZ

audit_logs
  id UUID PK
  project_id UUID → projects(id)
  tenant_id UUID → tenants(id)
  user_id UUID → users(id)
  action TEXT
  resource TEXT
  details JSONB
  created_at TIMESTAMPTZ
```

---

## Middleware Chain

```
logging → user auth → admin auth → tenant → router
```

| Middleware | Source | Effect |
|------------|--------|--------|
| Logging | `logging.go` | Logs method, path, status, duration via slog |
| Auth (user) | `auth.go` | Validates JWT against `JWT_SECRET`, injects user claims. **Skips** `/api/admin/*` and public routes. |
| Admin Auth | `admin.go` | Validates JWT against `ADMIN_JWT_SECRET`, injects admin claims. Skips `/api/admin/setup` and `/api/admin/login`. |
| Tenant | `tenant.go` | If `X-Tenant-ID` present, validates access via `user_tenants` table, injects tenant_id into context. Skips if header absent. |

### Public Routes (skip all auth)

- `GET /health`
- `POST /api/auth/signup`
- `POST /api/auth/login`
- `POST /api/admin/setup` (only works once)
- `POST /api/admin/login`

---

## API Routes

### Health

| Method | Path | Handler |
|--------|------|---------|
| GET | `/health` | `handlers.Health` |

### SaaS User Auth

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/auth/signup` | `AuthHandler.Signup` |
| POST | `/api/auth/login` | `AuthHandler.Login` |

### Platform Admin

| Method | Path | Handler | Auth |
|--------|------|---------|------|
| POST | `/api/admin/setup` | `AdminHandler.Setup` | None (only works once) |
| POST | `/api/admin/login` | `AdminHandler.Login` | None |
| POST | `/api/admin/projects` | `AdminHandler.CreateProject` | Admin JWT |
| GET | `/api/admin/projects` | `AdminHandler.ListProjects` | Admin JWT |
| POST | `/api/admin/projects/{pid}/roles` | `AdminHandler.AssignRole` | Admin JWT (super_admin only) |
| GET | `/api/admin/projects/{pid}/roles` | `AdminHandler.ListRoles` | Admin JWT (super_admin only) |
| DELETE | `/api/admin/projects/{pid}/roles/{aid}` | `AdminHandler.RemoveRole` | Admin JWT (super_admin only) |

### Tenants

| Method | Path | Handler |
|--------|------|---------|
| POST | `/api/projects/{pid}/tenants` | `TenantHandler.Create` |
| GET | `/api/projects/{pid}/tenants` | `TenantHandler.List` |
| GET | `/api/projects/{pid}/tenants/{tid}` | `TenantHandler.Get` |
| DELETE | `/api/projects/{pid}/tenants/{tid}` | `TenantHandler.Delete` |
| POST | `/api/projects/{pid}/tenants/{tid}/users` | `TenantHandler.AssignUser` |
| DELETE | `/api/projects/{pid}/tenants/{tid}/users/{uid}` | `TenantHandler.RemoveUser` |

---

## Bootstrap Flow

1. Server starts with `DATABASE_URL`, `JWT_SECRET`, `ADMIN_JWT_SECRET` env vars
2. Auto-migration runs (all `.up.sql` files in `db/migrations/`)
3. **Only on first ever startup:** `POST /api/admin/setup` creates:
   - First `admin_user` with `role = 'super_admin'`
   - First "Default Project" with `admin_owner_id` pointing to that admin
   - Returns JWT token + admin + project
4. Admin can now create additional projects via `POST /api/admin/projects`
5. Admin can assign other admin users to projects via `POST /api/admin/projects/{pid}/roles`
6. SaaS users can sign up via `POST /api/auth/signup` with a valid `project_id`
7. SaaS users are assigned to tenants via `POST /api/projects/{pid}/tenants/{tid}/users`

---

## Security Principles

- **Two JWT secrets**: `JWT_SECRET` for end-users, `ADMIN_JWT_SECRET` for platform admins
- **No trust**: permissions queried from DB on every request, never embedded in tokens
- **Tenant isolation**: every data query includes `project_id` and `tenant_id`
- **Fail fast**: missing env vars cause immediate startup failure
- **Passwords**: hashed with bcrypt (default cost)
- **API keys** (planned): hashed before storage
- **First-setup guard**: `POST /api/admin/setup` checks `admin_users` count — fails if >0

---

## Dependencies

| Library | Purpose |
|---------|---------|
| `pgx/v5` | PostgreSQL driver with connection pooling |
| `golang-migrate/migrate/v4` | Database migrations (file source + postgres driver) |
| `golang-jwt/jwt/v5` | JWT signing and validation (HS256) |
| `golang.org/x/crypto` | bcrypt password hashing |
