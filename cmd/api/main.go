package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/api"
	"github.com/UtopikCode/quickspaces-control-plane/application"
	"github.com/UtopikCode/quickspaces-control-plane/config"
	"github.com/UtopikCode/quickspaces-control-plane/docs"
	"github.com/UtopikCode/quickspaces-control-plane/execution"
	executionAdapters "github.com/UtopikCode/quickspaces-control-plane/execution/adapters"
	"github.com/UtopikCode/quickspaces-control-plane/persistence/postgres"
	contracts "github.com/UtopikCode/quickspaces-execution-contracts"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	var adapter contracts.ExecutionAdapter
	switch strings.ToLower(strings.TrimSpace(cfg.ExecutionProvider)) {
	case "aws":
		adapter = executionAdapters.NewAWSExecutionAdapter()
	case "truenas":
		adapter = executionAdapters.NewLocalExecutionAdapter()
	default:
		log.Fatalf("unsupported execution provider: %s", cfg.ExecutionProvider)
	}

	docs.SwaggerInfo.BasePath = "/api/v1"

	execSvc := execution.NewExecutionService(adapter)
	repo := postgres.NewWorkspaceRepository(db)
	service := application.NewWorkspaceService(repo, execSvc)
	handler := api.NewHandler(service)
	router := api.NewRouter(handler)

	server := &http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("starting control plane on %s", cfg.ListenAddress)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
