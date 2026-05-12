package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWeeklyPacksMigrationContract(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}

	candidates := []string{
		filepath.Join(wd, "migrations", "011_create_weekly_packs.up.sql"),
		filepath.Join(wd, "..", "..", "migrations", "011_create_weekly_packs.up.sql"),
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
		"CREATE TABLE IF NOT EXISTS weekly_packs",
		"created_by_user_id",
		"weekly_packs_status_check",
		"CREATE TABLE IF NOT EXISTS weekly_pack_movies",
		"sort_order",
		"weekly_pack_movies_pack_film_unique",
		"CREATE TABLE IF NOT EXISTS weekly_pack_votes",
		"weekly_pack_votes_pack_movie_fk",
		"weekly_pack_votes_vote_value_check",
		"weekly_pack_votes_limit_slot_check",
		"weekly_pack_votes_user_movie_unique",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_weekly_pack_votes_pack_user_value_slot",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(sql, fragment) {
			t.Fatalf("expected migration to contain %q", fragment)
		}
	}
}
