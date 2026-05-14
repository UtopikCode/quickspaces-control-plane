# Custom devcontainer image for QuickSpaces Control Plane
FROM mcr.microsoft.com/devcontainers/go:1-bookworm

ENV DEBIAN_FRONTEND=noninteractive
ENV GO111MODULE=on
ENV GOPATH=/go
ENV GOMODCACHE=/go/pkg/mod
ENV GOBIN=/usr/local/bin

# Prepare Go cache and install extra packages and developer tooling in one layer for better cache efficiency
RUN mkdir -p /go/pkg/mod /go/bin \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        postgresql-client \
    && rm -rf /var/lib/apt/lists/* \
    && go install github.com/swaggo/swag/cmd/swag@latest \
    && chown -R vscode:vscode /go
