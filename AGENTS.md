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