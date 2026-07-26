// Package db holds the shared Postgres connection (migration-time scaffolding;
// idiomatic Go reorg comes at the end). Go connects to the SAME database the Java
// stack uses — the parity harness runs both against one Postgres.
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // database/sql driver "postgres"
)

// Open connects to Postgres from OE_DB_* env vars. Defaults match the dev stack
// (docker-compose): db "clinlims", user postgres/admin. On the host the DB is
// published at localhost:15432; in a container it is the db service on :5432.
func Open() (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("OE_DB_HOST", "localhost"),
		env("OE_DB_PORT", "5432"),
		env("OE_DB_USER", "postgres"),
		env("OE_DB_PASSWORD", "admin"),
		env("OE_DB_NAME", "clinlims"),
		env("OE_DB_SSLMODE", "disable"),
	)
	database, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(5)
	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}
	return database, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
