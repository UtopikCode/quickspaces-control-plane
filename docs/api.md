# API Documentation

This repository uses `swaggo` to generate Swagger documentation from Go annotations.

## How it works

- API handlers are annotated with Swagger comments in `api/handlers.go`.
- The documentation generator reads those annotations and emits the Swagger spec in `docs/`.
- The control plane serves the Swagger UI at runtime via `http-swagger`.

## Regeneration

After changing or adding API handlers, run:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make generate
```

Or run directly:

```bash
go generate ./cmd/api
```

This command regenerates documentation in the `docs/` directory.

## Serving

Run the API server as usual:

```bash
go run ./cmd/api
```

Then open:

```text
http://localhost:8080/swagger/index.html
```

## Annotation guidelines

- Add `@Summary`, `@Description`, and `@Tags` for each endpoint.
- Use `@Accept json` and `@Produce json` when handling JSON input/output.
- Document path parameters with `@Param id path string true "Workspace ID"`.
- Use `@Success` and `@Failure` responses with concrete response types.
- Do not expose internal fields or infrastructure-only types.

## CI validation

The CI workflow validates that generated docs are up to date, and rejects changes when the `docs/` output is stale.
