// Command migrate is a standalone, opt-in goose CLI for the goose migration
// set under db/migrations. It is NOT invoked by cmd/openelis (the main
// server) and nothing in this repo runs it automatically.
//
// Why standalone: during coexistence, Java + Liquibase is the single schema
// owner (see migration/liquibase-to-goose-plan.md — "Go does NOT touch
// schema"). This binary exists so the migration set can be validated (e.g.
// applied to a scratch database to prove it is well-formed) without wiring
// any schema authority into the Go server itself. It becomes the real
// deployment tool only after the cutover procedure in that doc's § 6 runs.
//
// Usage:
//
//	go run ./cmd/migrate -dsn "postgres://user:pass@host:port/db?sslmode=disable" <goose-command> [args]
//
// <goose-command> is any normal goose subcommand: up, up-to, down, status,
// version, redo, reset, validate. See https://pressly.github.io/goose/.
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"openelis-go/db"
)

func main() {
	dsn := flag.String("dsn", "", "Postgres connection string (postgres://user:pass@host:port/db?sslmode=disable)")
	flag.Parse()

	args := flag.Args()
	if *dsn == "" || len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: migrate -dsn <conn-string> <goose-command> [args]")
		fmt.Fprintln(os.Stderr, "  e.g.: migrate -dsn \"$DSN\" status")
		os.Exit(2)
	}
	command, cmdArgs := args[0], args[1:]

	conn, err := sql.Open("postgres", *dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	// Liquibase's own extracted SQL leaves plenty of unqualified table
	// references (e.g. a FK to "data_export_task", not "clinlims.
	// data_export_task") — confirmed by running this against a clean DB:
	// they fail with "relation ... does not exist" unless search_path
	// resolves clinlims first. In production this works implicitly because
	// Java connects AS the "clinlims" role, and Postgres's default
	// search_path ("$user", public) resolves "$user" to a same-named
	// schema — clinlims happens to be both the role and the schema name.
	// Setting it explicitly here doesn't depend on which role -dsn uses.
	// SetMaxOpenConns(1) is required for this to stick: database/sql pools
	// connections, and SET on one pooled connection would silently not
	// apply to whichever different physical connection goose's next query
	// happens to borrow. This is a one-shot CLI, not a server, so pinning
	// to a single connection for the whole run has no real cost.
	conn.SetMaxOpenConns(1)
	if _, err := conn.Exec("SET search_path TO clinlims, public"); err != nil {
		log.Fatalf("set search_path: %v", err)
	}

	goose.SetBaseFS(db.Migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set dialect: %v", err)
	}

	// "migrations", not "db/migrations": go:embed paths in db/embed.go are
	// relative to that file's own directory (db/), so the embedded FS's
	// internal root is migrations/, not db/migrations/.
	if err := goose.RunWithOptions(command, conn, "migrations", cmdArgs); err != nil {
		log.Fatalf("goose %s: %v", command, err)
	}
}
