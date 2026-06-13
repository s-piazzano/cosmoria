# Cosmoria Folder Structure

This document explains the structure of the Cosmoria repository and the responsibility of each directory.

Cosmoria is designed as a modular backend engine built with Go, focused on multi-tenant SaaS applications.

---

# 🏗️ Root Structure Overview

cosmoria/
├── cmd/
├── internal/
├── pkg/
├── configs/
├── docs/
├── scripts/
├── docker/
├── go.mod
└── README.md

# 🚀 cmd/

Entry point of the application.

- Contains the main executable
- Responsible for bootstrapping the server
- Loads configuration and initializes dependencies

Example:

cmd/cosmoria/main.go


---

# 🧠 internal/

Core business logic of Cosmoria.

This is the most important part of the system.

It includes:

- API layer
- authentication
- multi-tenancy
- collections engine
- records system
- storage integration
- RBAC system

### Important rule:
Code inside `internal/` cannot be imported externally.

---

## internal/api/

Handles HTTP layer responsibilities:

- routing
- middleware execution
- request validation
- response formatting

No business logic should be implemented here.

---

## internal/core/

Core engine of Cosmoria.

Responsible for:
- application bootstrap
- shared context (project_id, tenant_id)
- configuration management

---

## internal/auth/

Authentication system:

- user login & signup
- JWT handling
- password hashing
- session management

---

## internal/tenant/

Multi-tenancy layer:

- tenant resolution
- tenant isolation logic
- request scoping

---

## internal/rbac/

Role-Based Access Control system:

- role definitions
- permission checks
- authorization rules

---

## internal/collections/

Dynamic schema system (low-code engine):

- collection definitions
- schema management (JSONB)
- API generation logic

---

## internal/records/

Generic data layer:

- CRUD operations
- JSONB-based storage
- filtering by project and tenant

---

## internal/storage/

File storage abstraction layer:

- S3-compatible integration
- file upload/download
- storage providers support

---

## internal/db/

Database layer:

- PostgreSQL connection
- migrations
- query helpers

---

## internal/realtime/

Real-time system (optional in MVP):

- WebSocket server
- event broadcasting
- database event subscriptions

---

## internal/audit/

Audit logging system:

- security logs
- user actions tracking
- system events recording

---

# 📦 pkg/

Reusable utility packages that can be safely shared across modules.

Examples:

- error handling
- validators
- helper functions

---

# ⚙️ configs/

Configuration files and environment handling:

- environment variables
- YAML config files
- configuration loader

---

# 📚 docs/

Project documentation:

- architecture overview
- API documentation
- folder structure (this file)
- design decisions

---

# 🧪 scripts/

Utility scripts for development:

- database migrations
- seeding data
- automation scripts

---

# 🐳 docker/

Containerization setup:

- Dockerfile
- docker-compose configuration

Used for local development and deployment.

---

# 🎯 Design Principles

Cosmoria follows these architectural rules:

- Separation of concerns
- Multi-tenancy enforced at backend level
- Minimal external dependencies
- PostgreSQL as the primary data source
- JSONB-based flexible schemas

---

# 🚧 Summary

The Cosmoria codebase is structured to behave like an infrastructure engine, not a traditional web application.

Each module has a strict responsibility to ensure scalability, maintainability, and clarity for contributors.