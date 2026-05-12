# Architecture

## Control Plane Purpose

The Control Plane is a stateless API that defines desired workspace state and delegates execution to ExecutionAdapters.
It does not perform infrastructure actions, background processing, or reconciliation loops.

## Core responsibilities

- accept workspace lifecycle requests
- persist desired and actual state in PostgreSQL
- delegate runtime actions through `ExecutionAdapter`
- expose a clean HTTP interface

## Separation of concerns

- `/api` contains HTTP handlers and request routing
- `/application` contains use cases and business rules
- `/domain` defines the workspace model and errors
- `/persistence` contains the Postgres repository implementation
- `/execution` contains adapter wiring and the adapter interface
- `/cmd/api` contains the application entrypoint

## Stateless behavior

The Control Plane never performs polling, retries, or background jobs.
Each request is handled synchronously:

- `POST /api/v1/workspaces/{id}/start`
  - sets `desiredState = "running"`
  - calls `ExecutionAdapter.StartWorkspace`
- `POST /api/v1/workspaces/{id}/stop`
  - sets `desiredState = "stopped"`
  - calls `ExecutionAdapter.StopWorkspace`
- `POST /api/v1/workspaces/{id}/reconcile`
  - calls `ExecutionAdapter.GetWorkspaceStatus`
  - updates `actualState`

## ExecutionAdapter role

ExecutionAdapters are the only integration boundary for runtime intent.
The Control Plane does not contain AWS, Docker, or Kubernetes logic.

## System diagram

Client -> HTTP API -> Application Service -> Repository / ExecutionAdapter

The ExecutionAdapter is injected by configuration, preserving portability.
