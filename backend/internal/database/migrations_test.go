package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveMigrationsPath(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	migrationsDir := filepath.Join(wd, "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	path, err := resolveMigrationsPath()
	if err != nil {
		t.Fatalf("resolveMigrationsPath failed: %v", err)
	}
	if !strings.HasPrefix(path, "file://") {
		t.Fatalf("expected file:// prefix, got %s", path)
	}
	if !strings.Contains(path, "/migrations") {
		t.Fatalf("expected migrations suffix, got %s", path)
	}
}
