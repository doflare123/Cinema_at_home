package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTelegramNotificationsMigrationContract(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	candidates := []string{
		filepath.Join(wd, "migrations", "015_create_telegram_notifications.up.sql"),
		filepath.Join(wd, "..", "..", "migrations", "015_create_telegram_notifications.up.sql"),
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
		"CREATE TABLE IF NOT EXISTS telegram_notifications",
		"type VARCHAR(32) NOT NULL",
		"status VARCHAR(16) NOT NULL DEFAULT 'queued'",
		"target_user_id INT NULL REFERENCES users(id) ON DELETE SET NULL",
		"payload JSONB NOT NULL DEFAULT '{}'::jsonb",
		"enqueued_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT",
		"telegram_notifications_type_check",
		"telegram_notifications_status_check",
		"telegram_notifications_attempts_check",
		"CREATE INDEX IF NOT EXISTS idx_telegram_notifications_status_created",
		"CREATE INDEX IF NOT EXISTS idx_telegram_notifications_type_created",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("expected migration to contain %q", fragment)
		}
	}
}
