package database

import "testing"

func TestReviewsMigrationPostgresIntegration(t *testing.T) {
	tx, _ := openPostgresIntegrationTx(t)

	if _, err := tx.Exec(`
		CREATE TABLE users (id SERIAL PRIMARY KEY);
		CREATE TABLE films (id SERIAL PRIMARY KEY);
	`); err != nil {
		t.Fatalf("create dependencies failed: %v", err)
	}

	sqlText := readMigrationSQLForIntegration(t, "012_create_reviews.up.sql")
	if _, err := tx.Exec(sqlText); err != nil {
		t.Fatalf("apply reviews migration failed: %v", err)
	}

	var userID, filmID int
	if err := tx.QueryRow(`INSERT INTO users DEFAULT VALUES RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("insert user failed: %v", err)
	}
	if err := tx.QueryRow(`INSERT INTO films DEFAULT VALUES RETURNING id`).Scan(&filmID); err != nil {
		t.Fatalf("insert film failed: %v", err)
	}

	var simpleFinal float64
	if err := tx.QueryRow(`
		INSERT INTO reviews (film_id, user_id, mode, score, final_score, criteria_scores)
		VALUES ($1, $2, 'simple', 7, 1, '{}'::jsonb)
		RETURNING final_score
	`, filmID, userID).Scan(&simpleFinal); err != nil {
		t.Fatalf("insert simple review failed: %v", err)
	}
	if simpleFinal != 7 {
		t.Fatalf("expected final_score=7, got %v", simpleFinal)
	}

	var userID2 int
	if err := tx.QueryRow(`INSERT INTO users DEFAULT VALUES RETURNING id`).Scan(&userID2); err != nil {
		t.Fatalf("insert second user failed: %v", err)
	}

	var criteriaFinal float64
	if err := tx.QueryRow(`
		INSERT INTO reviews (film_id, user_id, mode, final_score, criteria_scores)
		VALUES ($1, $2, 'criteria', 1, '{"plot":8,"visual":10}'::jsonb)
		RETURNING final_score
	`, filmID, userID2).Scan(&criteriaFinal); err != nil {
		t.Fatalf("insert criteria review failed: %v", err)
	}
	if criteriaFinal != 9 {
		t.Fatalf("expected final_score=9, got %v", criteriaFinal)
	}
}
