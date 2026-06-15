# Cosmoria — Collections Engine

Collections are dynamic schemas stored as JSONB in PostgreSQL.  
They allow runtime-definable data structures without SQL schema migrations.

---

## Schema Format

A collection schema is defined as:

```json
{
  "fields": [
    { "name": "title",   "type": "string", "required": true },
    { "name": "price",   "type": "number", "required": false },
    { "name": "active",  "type": "boolean", "required": false }
  ]
}
```

### Field Types

| Type | Go type | JSON example |
|------|---------|-------------|
| `string` | `string` | `"hello"` |
| `number` | `float64` | `42.5` |
| `boolean` | `bool` | `true` |

### Validation Rules

On every record create/update, the engine validates:

1. **Required fields** — if `required: true`, the field must be present and non-nil
2. **Type check** — the value must match the declared type:
   - `string` → JSON string
   - `number` → JSON number (stored as `float64`)
   - `boolean` → JSON boolean

Fields not defined in the schema are allowed and stored as-is in the JSONB.

> Schema mutation (admin) is always permitted. Existing records are NOT re-validated after a schema change.

---

## Records CRUD

Records are tenant-scoped JSONB entities.

### Create

```
POST /api/projects/{pid}/tenants/{tid}/collections/{cid}/records
Body: { "data": { "title": "Hello", "price": 10 } }
```

Returns the created record with `id` and `created_at`.

### Read

```
GET /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}
```

Returns a single record object.

### List (cursor pagination)

```
GET /api/projects/{pid}/tenants/{tid}/collections/{cid}/records
  ?cursor=<record_id>&limit=50
```

Response:
```json
{
  "data": [ /* records */ ],
  "next_cursor": "uuid"  // omitted when no more pages
}
```

Pagination is ordered by `(created_at, id)`. The `cursor` parameter is the `id` of the last record from the previous page.

### Update (full replacement)

```
PUT /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}
Body: { "data": { "title": "Updated", "price": 20 } }
```

The entire `data` JSONB is replaced (not merged). Validated against the schema.

### Delete

```
DELETE /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid}
```

Returns `204 No Content`.

---

## Architecture

```
┌─────────────────────────────────────────────┐
│  Records Handler (user-facing, RBAC-wrapped)│
├─────────────────────────────────────────────┤
│  Records Service                            │
│  - CRUD with cursor pagination              │
│  - Schema validation via Collection Service │
├─────────────────────────────────────────────┤
│  Collections Handler (admin-only)           │
├─────────────────────────────────────────────┤
│  Collections Service                        │
│  - CRUD, Schema mutation                    │
│  - ValidateData(string|number|boolean)      │
├─────────────────────────────────────────────┤
│  PostgreSQL (JSONB columns)                  │
└─────────────────────────────────────────────┘
```
