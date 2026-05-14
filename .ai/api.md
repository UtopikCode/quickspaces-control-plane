# API Versioning Policy

- ALWAYS use `/api/v1` prefix for every API endpoint.
- NEVER create unversioned endpoints under `/workspaces`, `/auth`, or workspace actions.
- NEVER mix version prefixes such as `/api/workspaces`, `/api/v2/workspaces`, or `/workspaces` with `/api/v1`.
- NEVER hardcode the version in handler logic.
- ALWAYS rely on router grouping to apply the API version prefix.
- ALWAYS keep route annotations relative to the API group, for example `/workspaces` instead of `/api/v1/workspaces`.
- ALWAYS update Swagger `BasePath` to `/api/v1` when route versions change.
- ALWAYS add new major versions under `/api/v2`, `/api/v3`, etc. and preserve `/api/v1` behavior for existing clients.
