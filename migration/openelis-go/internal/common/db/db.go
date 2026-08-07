// Package db holds the shared Postgres connection (migration-time scaffolding;
// idiomatic Go reorg comes at the end). Go connects to the SAME database the Java
// stack uses — the parity harness runs both against one Postgres.
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // database/sql driver "postgres" (kept for a2 domains)
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenGORM opens a GORM *gorm.DB backed by the clinlims Postgres instance.
// b1 domains (dictionary-categories, uom, test-catalog, TestCatalog) use GORM.
// a2 domains (localization, status-types) extract *sql.DB via gormDB.DB().
func OpenGORM() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("OE_DB_HOST", "localhost"),
		env("OE_DB_PORT", "5432"),
		env("OE_DB_USER", "postgres"),
		env("OE_DB_PASSWORD", "admin"),
		env("OE_DB_NAME", "clinlims"),
		env("OE_DB_SSLMODE", "disable"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Silence GORM's default logger in production; enable with OE_DB_LOG=true.
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(5)
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// Open keeps the old *sql.DB interface for callers that haven't migrated to
// GORM yet (currently none in b1; retained for backward compatibility).
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
