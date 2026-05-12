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

## Running locally

1. Configure PostgreSQL and export `DATABASE_URL`.
2. Optionally set `EXECUTION_PROVIDER` to `aws` or `truenas`.
3. Run the server:

```bash
go run ./cmd/api
```

By default the server listens on `:8080` unless `LISTEN_ADDR` is provided.

## Development commands

Use the Makefile to run tests, linting, and formatting:

```bash
make test
make lint
make format
make ci
```

`make ci` will verify formatting, run static analysis, and execute the full test suite.

## Environment variables

- `DATABASE_URL` — Postgres DSN, required
- `EXECUTION_PROVIDER` — execution adapter provider (`aws` or `truenas`), defaults to `truenas`
- `LISTEN_ADDR` — HTTP listen address, defaults to `:8080`

## Database schema

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

## Notes

- The Control Plane defines desired state and delegates execution.
- It never performs orchestration, retries, or background reconciliation.
- Adapter implementations are intentionally separate from control plane behavior.
