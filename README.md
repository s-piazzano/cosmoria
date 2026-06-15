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
- **MCP server** — `cosmoria mcp` lets Claude Code / Cursor interact directly via JSON-RPC

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
- Zero-config startup (sensible defaults, generated JWT secrets)

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
