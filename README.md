# quickspaces-control-plane

QuickSpaces Control Plane is a stateless API for managing workspace desired state and delegating execution to external adapters.

## What is the Control Plane?

This repository implements a clean control plane for QuickSpaces:

- exposes workspace lifecycle APIs
- remains fully stateless
- does not contain AWS, Docker, or Kubernetes logic
- delegates execution to `ExecutionAdapter`
- persists desired and actual workspace state in PostgreSQL

## Architecture

- `cmd/api` — entrypoint
- `api` — HTTP handlers and router
- `application` — use cases and business services
- `domain` — workspace model and domain errors
- `persistence` — PostgreSQL repository
- `execution` — execution adapter wiring
- `config` — environment-driven configuration
- `docs` — architecture documentation

## API Versioning

All routes are explicitly versioned under `/api/v1`.

- Versioned API paths use the stable prefix `/api/v1`.
- Non-API routes such as `/swagger` and `/health` remain outside the API version prefix.
- Future major versions such as `/api/v2` can be added without changing existing v1 route handlers.

## Running locally

1. Configure PostgreSQL and export `DATABASE_URL`.
2. Optionally set `EXECUTION_PROVIDER` to `aws` or `truenas`.
3. Run the server:

```bash
go run ./cmd/api
```

By default the server listens on `:8080` unless `LISTEN_ADDR` is provided.

## Development commands

Use the Makefile to run tests, linting, generation, formatting, and database initialization:

```bash
make test
make lint
make vet
make format
make check-format
make generate-swagger
make generate-ent
make migrate-ent
make generate-ent-ddl
make init-db
make ci
```

Run generator directly with:

```bash
go generate ./cmd/api
```

`make ci` will verify formatting, run static analysis, and execute the full test suite.

## Environment variables

- `DATABASE_URL` — Postgres DSN, required
- `EXECUTION_PROVIDER` — execution adapter provider (`aws` or `truenas`), defaults to `truenas`
- `LISTEN_ADDR` — HTTP listen address, defaults to `:8080`
- `GITHUB_CLIENT_ID` — GitHub OAuth App client ID, required
- `GITHUB_CLIENT_SECRET` — GitHub OAuth App client secret, required
- `GITHUB_REDIRECT_URL` — OAuth callback URL, required
- `ADMIN_USERS` — comma-separated bootstrap access rule specs used when no `access_rules` exist, optional. Valid forms include `alice`, `user:alice`, `org:acme`, or `team:acme/developers`.

If you want the Swagger UI authorize button to work, register the backend callback URL as the OAuth redirect URL:

- `http://localhost:8080/api/v1/auth/callback`

## Authentication and authorization

This control plane is stateless. GitHub is the source of truth for identity, and the database stores only access rules.

- `GET /api/v1/auth/login` starts GitHub OAuth.
- `GET /api/v1/auth/callback?code=...` returns an OAuth access token and the authenticated identity.
- All workspace routes require `Authorization: Bearer <token>`.
- Access is granted by `access_rules` in the database, not by internal user records.

## Database schema

Initialize the database schema with:

```bash
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=disable make init-db
```

Or create the tables manually from `db/schema.sql`.

Create the `workspaces` table before running the API, for example:

```sql
CREATE TABLE IF NOT EXISTS workspaces (
    id TEXT PRIMARY KEY,
    repo TEXT NOT NULL,
    owner TEXT NOT NULL,
    ref TEXT NOT NULL,
    desired_state TEXT NOT NULL,
    actual_state TEXT NOT NULL,
    execution_profile JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

Create the `access_rules` table to drive authorization:

```sql
CREATE TABLE IF NOT EXISTS access_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "type" TEXT NOT NULL,
    value TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
```

Supported rule types:

- `user` with `value` set to a GitHub login
- `org` with `value` set to a GitHub organization login
- `team` with `value` set to `org/team`

## API

### Create workspace

```http
POST /api/v1/workspaces
Content-Type: application/json

{
  "repo": "github.com/example/repo",
  "owner": "team-a",
  "ref": "main",
  "executionProfile": {
    "provider": "aws",
    "backend": "ecs"
  }
}
```

### List workspaces

```http
GET /api/v1/workspaces
```

### Get a workspace

```http
GET /api/v1/workspaces/{id}
```

### Start a workspace

```http
POST /api/v1/workspaces/{id}/start
```

### Stop a workspace

```http
POST /api/v1/workspaces/{id}/stop
```

### Reconcile a workspace

```http
POST /api/v1/workspaces/{id}/reconcile
```

### Health

```http
GET /api/v1/health
```

## API Documentation

Swagger UI is available after the server starts:

- `http://localhost:8080/swagger/index.html`

The documentation is generated from code annotations and committed under `docs/`.

To regenerate docs after API changes:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make generate
```

## Notes

- The Control Plane defines desired state and delegates execution.
- It never performs orchestration, retries, or background reconciliation.
- Adapter implementations are intentionally separate from control plane behavior.
