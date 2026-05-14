package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/UtopikCode/quickspaces-control-plane/ent"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	flag.StringVar(&dsn, "db", dsn, "Postgres DSN")
	flag.Parse()

	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed to apply ent migrations: %v", err)
	}

	log.Println("ent migrations applied successfully")
}
