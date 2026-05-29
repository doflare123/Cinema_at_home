package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func openPostgresIntegrationTx(t *testing.T) (*sql.Tx, string) {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("TEST_POSTGRES_DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open postgres failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	schema := fmt.Sprintf("it_%d", time.Now().UnixNano())
	if _, err := db.Exec(fmt.Sprintf("CREATE SCHEMA %s", schema)); err != nil {
		t.Fatalf("create schema failed: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
	})

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx failed: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	if _, err := tx.Exec(fmt.Sprintf("SET search_path TO %s", schema)); err != nil {
		t.Fatalf("set search path failed: %v", err)
	}
	return tx, schema
}

func readMigrationSQLForIntegration(t *testing.T, file string) string {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	candidates := []string{
		filepath.Join(wd, "migrations", file),
		filepath.Join(wd, "..", "..", "migrations", file),
	}
	for _, candidate := range candidates {
		content, readErr := os.ReadFile(candidate)
		if readErr == nil {
			return string(content)
		}
	}
	t.Fatalf("migration file %s not found", file)
	return ""
}
