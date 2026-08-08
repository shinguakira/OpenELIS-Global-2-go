// Package db embeds the goose migration set (migrations/*.sql) so it ships
// inside the Go binary without needing a separate directory on disk at
// runtime. See cmd/migrate for the CLI that applies these, and
// migration/liquibase-to-goose-plan.md for how they were generated.
package db

import "embed"

// Migrations holds every goose migration file under migrations/.
//
//go:embed migrations/*.sql
var Migrations embed.FS
