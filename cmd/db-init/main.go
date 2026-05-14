package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/UtopikCode/quickspaces-control-plane/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	flag.StringVar(&dsn, "db", dsn, "MongoDB URI")
	flag.Parse()

	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DATABASE_URL"))
	}
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	databaseName := strings.TrimSpace(os.Getenv("DATABASE_NAME"))
	if databaseName == "" {
		databaseName = "quickspaces"
	}
	cfg := &config.Config{
		DatabaseURL:  dsn,
		DatabaseName: databaseName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.DatabaseURL))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("failed to disconnect database: %v", err)
		}
	}()

	migrationsDir, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatalf("failed to resolve migrations directory: %v", err)
	}

	driver, err := mongodb.WithInstance(client, &mongodb.Config{DatabaseName: cfg.DatabaseName})
	if err != nil {
		log.Fatalf("failed to create MongoDB migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsDir),
		"mongodb",
		driver,
	)
	if err != nil {
		log.Fatalf("failed to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("MongoDB migrations applied successfully")
}
