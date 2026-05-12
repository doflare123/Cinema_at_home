package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewsMigrationContract(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	candidates := []string{
		filepath.Join(wd, "migrations", "012_create_reviews.up.sql"),
		filepath.Join(wd, "..", "..", "migrations", "012_create_reviews.up.sql"),
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
		"CREATE TABLE IF NOT EXISTS reviews",
		"mode VARCHAR(16) NOT NULL",
		"score SMALLINT NULL",
		"final_score NUMERIC(4,2) NOT NULL",
		"criteria_scores JSONB NOT NULL DEFAULT '{}'::jsonb",
		"comment TEXT NOT NULL DEFAULT ''",
		"reviews_mode_check",
		"reviews_score_check",
		"reviews_final_score_check",
		"reviews_user_film_unique",
		"CREATE OR REPLACE FUNCTION set_review_final_score()",
		"NEW.final_score := NEW.score",
		"NEW.final_score := ROUND(total / score_count, 2)",
		"CREATE TRIGGER reviews_set_final_score",
		"BEFORE INSERT OR UPDATE OF mode, score, final_score, criteria_scores, comment",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("expected migration to contain %q", fragment)
		}
	}

	forbiddenFragments := []string{
		"override_score",
	}

	for _, fragment := range forbiddenFragments {
		if strings.Contains(sql, fragment) {
			t.Fatalf("expected migration to forbid manual aggregate override fragment %q", fragment)
		}
	}
}
