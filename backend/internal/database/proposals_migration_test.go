package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMovieProposalsMigrationContract(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	candidates := []string{
		filepath.Join(wd, "migrations", "013_create_movie_proposals.up.sql"),
		filepath.Join(wd, "..", "..", "migrations", "013_create_movie_proposals.up.sql"),
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
		"CREATE TABLE IF NOT EXISTS movie_proposals",
		"proposed_by_user_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT",
		"moderated_by_user_id INT NULL REFERENCES users(id) ON DELETE RESTRICT",
		"moderated_at TIMESTAMP NULL",
		"film_id INT NULL REFERENCES films(id) ON DELETE RESTRICT",
		"movie_proposals_source_check",
		"movie_proposals_status_check",
		"movie_proposals_moderation_audit_check",
		"CREATE INDEX IF NOT EXISTS idx_movie_proposals_status_created",
		"ON movie_proposals(status, created_at DESC, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_movie_proposals_proposed_by_created",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("expected migration to contain %q", fragment)
		}
	}
}
