// @title QuickSpaces Control Plane API
// @version 1.0
// @description Stateless Control Plane API for QuickSpaces.
// @host localhost:8080
// @BasePath /api/v1
// @schemes http
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/api"
	"github.com/UtopikCode/quickspaces-control-plane/application"
	"github.com/UtopikCode/quickspaces-control-plane/config"
	"github.com/UtopikCode/quickspaces-control-plane/execution"
	"github.com/UtopikCode/quickspaces-control-plane/internal/application/auth"
	githubclient "github.com/UtopikCode/quickspaces-control-plane/internal/infrastructure/github"
	mongopersistence "github.com/UtopikCode/quickspaces-control-plane/persistence/mongo"
	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
	truenasadapter "github.com/UtopikCode/quickspaces-execution-truenas/adapter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// openAPISpec is a minimal OpenAPI 3.0.0 spec for the Control Plane API

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DatabaseURL))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("failed to disconnect mongo client: %v", err)
		}
	}()

	db := client.Database(cfg.DatabaseName)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	registry := execution.NewAdapterRegistry()
	// Register execution adapters via imported provider packages.
	// The control-plane core itself remains adapter-agnostic and resolves providers by name at runtime.
	registry.Register("truenas", func(hostConfig json.RawMessage) (contracts.ExecutionAdapter, error) {
		return truenasadapter.NewDefaultDockerExecutionAdapter()
	})

	repo := mongopersistence.NewWorkspaceRepository(db)
	hostRepo := mongopersistence.NewHostRepository(db)
	execSvc := execution.NewExecutionService(registry, hostRepo)
	accessRepo := mongopersistence.NewAccessRuleRepository(db)

	// Create GitHub client for OAuth
	githubClient := githubclient.NewClient(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.GitHubRedirectURL)

	service := application.NewWorkspaceService(repo, hostRepo, execSvc)
	hostService := application.NewHostService(hostRepo)
	authService := auth.NewService(accessRepo, cfg.InitialAccessRules)
	handler := api.NewHandler(service, hostService, authService, githubClient)
	apiRouter := api.NewRouter(handler)

	openAPISpecPath, err := resolveOpenAPISpecPath()
	if err != nil {
		log.Fatalf("failed to resolve openapi spec path: %v", err)
	}

	// Create root mux
	mux := http.NewServeMux()

	// Mount API routes
	mux.Handle("/api/", apiRouter)

	// Serve generated OpenAPI spec from docs/swagger.json
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		serveOpenAPISpec(w, r, openAPISpecPath)
	})

	// Mount Scalar UI (development mode only)
	scalarEnabled := strings.ToLower(strings.TrimSpace(os.Getenv("ENABLE_SCALAR"))) == "true"
	if scalarEnabled {
		// Serve Scalar HTML with OAuth pre-configured
		authConfig := strings.NewReplacer(
			"{{GITHUB_CLIENT_ID}}", cfg.GitHubClientID,
			"{{GITHUB_CLIENT_SECRET}}", cfg.GitHubClientSecret,
			"{{GITHUB_REDIRECT_URI}}", cfg.GitHubRedirectURL,
		).Replace(`                  'x-scalar-client-id': '{{GITHUB_CLIENT_ID}}',
                  clientSecret: '{{GITHUB_CLIENT_SECRET}}',
                  'x-scalar-redirect-uri': '{{GITHUB_REDIRECT_URI}}',
                  authorizationUrl: 'https://github.com/login/oauth/authorize',
                  tokenUrl: 'http://localhost:8080/api/v1/auth/token',
                  selectedScopes: ['read:org'],
                  'x-usePkce': 'SHA-256',`)

		scalarHTML := strings.ReplaceAll(`<!doctype html>
<html>
  <head>
    <title>Scalar - API Reference</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      html, body {
        margin: 0;
        padding: 0;
        min-height: 100vh;
        overflow: auto;
      }
      #app {
        position: absolute;
        top: 30px;
        bottom: 0;
        left: 0;
        right: 0;
        overflow: auto;
      }
    </style>
  </head>
  <body>
    <div id="app"></div>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
      Scalar.createApiReference('#app', {
        url: '/openapi.json',
        theme: 'default',
        hideDownloadButton: false,
        hideModels: false,
        layout: 'modern',
        authentication: {
          preferredSecurityScheme: 'GitHubOAuth',
          securitySchemes: {
            GitHubOAuth: {
              flows: {
                authorizationCode: {
{{AUTH_CONFIG}}
                },
              },
            },
          },
        },
      });
    </script>
  </body>
</html>`, "{{AUTH_CONFIG}}", authConfig)

		mux.HandleFunc("/scalar/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/scalar/" || r.URL.Path == "/scalar/index.html" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				if _, err := w.Write([]byte(scalarHTML)); err != nil {
					log.Printf("failed to write scalar html: %v", err)
				}
				return
			}
			http.NotFound(w, r)
		})

		mux.HandleFunc("/scalar", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/scalar/", http.StatusMovedPermanently)
		})
	}

	server := &http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("starting control plane on %s", cfg.ListenAddress)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func resolveOpenAPISpecPath() (string, error) {
	paths := []string{}

	if wd, err := os.Getwd(); err == nil {
		paths = append(paths, wd)
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		paths = append(paths, exeDir)
	}

	for _, base := range paths {
		if specPath, ok := findInParents(base, filepath.Join("docs", "swagger.json")); ok {
			return specPath, nil
		}
	}

	return "", fmt.Errorf("openapi spec not found")
}

func findInParents(start, rel string) (string, bool) {
	current := filepath.Clean(start)
	for {
		candidate := filepath.Join(current, rel)
		if _, err := os.Stat(candidate); err == nil {
			if abs, err := filepath.Abs(candidate); err == nil {
				return abs, true
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return "", false
}

func serveOpenAPISpec(w http.ResponseWriter, r *http.Request, specPath string) {
	w.Header().Set("Content-Type", "application/json")
	http.ServeFile(w, r, specPath)
}
