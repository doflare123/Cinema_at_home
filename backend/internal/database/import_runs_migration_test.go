package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportRunsMigrationContract(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	candidates := []string{
		filepath.Join(wd, "migrations", "014_create_import_runs.up.sql"),
		filepath.Join(wd, "..", "..", "migrations", "014_create_import_runs.up.sql"),
	}

	var (
		content []byte
		readErr error
	)
	for _, candidate := range candidates {
		content, readErr = os.ReadFile(candidate)
		if readErr == nil {
			break
		}
	}
	if readErr != nil {
		t.Fatalf("read migration failed: %v", readErr)
	}

	sql := string(content)
	requiredFragments := []string{
		"CREATE TABLE IF NOT EXISTS import_runs",
		"imported_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT",
		"dry_run BOOLEAN NOT NULL DEFAULT FALSE",
		"rows_created INT NOT NULL DEFAULT 0",
		"rows_skipped_existing INT NOT NULL DEFAULT 0",
		"rows_skipped_invalid INT NOT NULL DEFAULT 0",
		"CREATE INDEX IF NOT EXISTS idx_import_runs_created_at",
		"CREATE INDEX IF NOT EXISTS idx_import_runs_imported_by_created",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("expected migration to contain %q", fragment)
		}
	}
}
