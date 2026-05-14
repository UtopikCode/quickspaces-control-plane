# Swagger documentation guidance

- ALWAYS document new endpoints using swag annotations in Go source.
- NEVER modify generated files under `docs/` manually.
- ALWAYS run `swag init` or `make generate-swagger` after API changes.
- NEVER expose internal or infrastructure-only fields in Swagger schemas.
- API documentation must reflect the real behavior of the current API.
- Swagger exists only as documentation, not as a source of truth for runtime behavior.
