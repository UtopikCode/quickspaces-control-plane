# AI Instructions for quickspaces-control-plane

- Control Plane MUST remain stateless.
- NEVER add AWS, Docker, or Kubernetes logic in this repository.
- NEVER create background jobs, polling loops, retries, or reconciliation loops.
- ALWAYS delegate runtime actions to an ExecutionAdapter.
- ExecutionAdapter is the only integration surface for execution behavior.
- ExecutionProfile MUST be treated as opaque input and never interpreted by control plane logic.
- Control Plane defines desired state and delegates execution; it does NOT manage actual infrastructure.
- Keep business logic in application services and state in persistence.
