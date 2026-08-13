// Throwaway one-off: create a scratch DB + clinlims schema on the dev
// Postgres for goose-migration verification. Not part of the module build
// (leading underscore dir is ignored by `go build ./...`). Delete after use.
package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	admin, err := sql.Open("postgres", "postgres://postgres:admin@localhost:15432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer admin.Close()

	var exists bool
	err = admin.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = 'clinlims_goose_verify')").Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	if exists {
		log.Println("dropping existing clinlims_goose_verify")
		if _, err := admin.Exec("DROP DATABASE clinlims_goose_verify"); err != nil {
			log.Fatal(err)
		}
	}
	if _, err := admin.Exec("CREATE DATABASE clinlims_goose_verify"); err != nil {
		log.Fatal(err)
	}
	log.Println("created clinlims_goose_verify (empty — baseline dump creates the clinlims schema itself)")
}
