# Cosmoria

Written in Go. Single binary. Runs anywhere PostgreSQL runs.

A backend engine for multi-tenant SaaS applications. Define tenants,
collections, and permissions — Cosmoria provides the REST API, auth,
tenant isolation, and RBAC.

```dockerfile
docker run \
  -e DATABASE_URL=postgres://postgres:password@host:5432/cosmoria?sslmode=disable \
  -p 8080:8080 \
  ghcr.io/s-piazzano/cosmoria
```

---

## Define your app in one file

Write a `cosmoria.yaml`, then `cosmoria serve`. Everything is created
idempotently.

```yaml
# cosmoria.yaml
project: my-saas
tenants:
  - name: acme-corp
  - name: globex
collections:
  - name: posts
    schema:
      fields:
        - { name: title, type: string, required: true }
        - { name: body, type: string }
        - { name: published, type: boolean }
roles:
  - name: editor
    permissions:
      - { resource: records, action: create }
      - { resource: records, action: read }
      - { resource: records, action: update }
```

That's it. Users sign up, get assigned to tenants, and start creating
records. The API is live with Swagger UI at `/docs/`.

---

## Built for AI agents

Cosmoria exposes everything via:

- **REST API** — 33 endpoints, documented via OpenAPI (Swagger UI at `/docs/`)
- **Declarative config** — write YAML, the engine applies it idempotently
- **MCP server** — `cosmoria mcp` exposes the backend to Claude, Cursor, Windsurf, Continue.dev, OpenCode, and any MCP-compatible client

An AI agent writes a config file, runs `cosmoria serve`, and the
backend is ready in seconds.

---

## Features

- Multi-tenancy with project + tenant scoping
- Dynamic collections with JSONB schemas
- RBAC with per-project roles and permissions
- User and admin authentication (bcrypt + JWT)
- Cursor-paginated records API
- Auto-generated OpenAPI spec
- TypeScript SDK (zero runtime deps)
- CLI: `serve`, `dev` (hot reload), `migrate new/up/down`, `init`
- MCP server — compatible with Claude, Cursor, Windsurf, Continue.dev, OpenCode, and any MCP client
- Zero-config startup (sensible defaults, generated JWT secrets)

---

## MCP Server

Cosmoria exposes its entire backend as an [MCP](https://modelcontextprotocol.io) server via JSON-RPC 2.0 over stdio.

```bash
cosmoria mcp
```

No authentication required — the MCP server connects directly to PostgreSQL and operates with superuser privileges. Intended for trusted local environments only.

### Configuration example (Claude Desktop)

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "cosmoria": {
      "command": "/usr/local/bin/cosmoria",
      "args": ["mcp"],
      "env": {
        "DATABASE_URL": "postgres://localhost:5432/cosmoria?sslmode=disable"
      }
    }
  }
}
```

### Available tools (21)

| Group | Tool | Description |
|-------|------|-------------|
| Setup | `cosmoria_setup` | Bootstrap super_admin and default project (once) |
| Projects | `cosmoria_project_create` | Create a new project |
| | `cosmoria_project_list` | List accessible projects |
| Tenants | `cosmoria_tenant_create` | Create a tenant |
| | `cosmoria_tenant_list` | List tenants |
| | `cosmoria_tenant_get` | Get a tenant by ID |
| Collections | `cosmoria_collection_create` | Create a collection with schema |
| | `cosmoria_collection_list` | List collections |
| | `cosmoria_collection_get` | Get collection schema |
| RBAC | `cosmoria_role_create` | Create a role |
| | `cosmoria_role_list` | List roles with permissions |
| | `cosmoria_role_set_permission` | Assign (resource, action) to role |
| | `cosmoria_role_list_permissions` | List role permissions |
| Records | `cosmoria_record_create` | Create a record |
| | `cosmoria_record_list` | List records (cursor pagination) |
| | `cosmoria_record_get` | Get a record |
| | `cosmoria_record_update` | Update a record |
| | `cosmoria_record_delete` | Delete a record |
| Users | `cosmoria_user_assign_role` | Assign RBAC role to user |
| Files | `cosmoria_file_list` | List files for a tenant |
| Audit | `cosmoria_audit_list` | List audit logs |

An AI agent can provision projects, tenants, collections, and roles entirely through natural language — no HTTP calls needed.

---

## Architecture

```
Your Application
        |
        v
   Cosmoria Engine  ──── MCP (agents)
        |
   ------------
   |          |
PostgreSQL   S3 (planned)
```

Single binary. Static linking. No runtime dependencies.

---

## Status

🚧 v0 — actively evolving, not production-ready yet.

## License

Apache License 2.0
