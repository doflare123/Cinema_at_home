package database

import "testing"

func TestProposalsMigrationPostgresIntegration(t *testing.T) {
	tx, _ := openPostgresIntegrationTx(t)

	if _, err := tx.Exec(`
		CREATE TABLE users (id SERIAL PRIMARY KEY);
		CREATE TABLE films (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL UNIQUE
		);
	`); err != nil {
		t.Fatalf("create dependencies failed: %v", err)
	}

	sqlText := readMigrationSQLForIntegration(t, "013_create_movie_proposals.up.sql")
	if _, err := tx.Exec(sqlText); err != nil {
		t.Fatalf("apply proposals migration failed: %v", err)
	}

	var proposerID, moderatorID, filmID int
	if err := tx.QueryRow(`INSERT INTO users DEFAULT VALUES RETURNING id`).Scan(&proposerID); err != nil {
		t.Fatalf("insert proposer failed: %v", err)
	}
	if err := tx.QueryRow(`INSERT INTO users DEFAULT VALUES RETURNING id`).Scan(&moderatorID); err != nil {
		t.Fatalf("insert moderator failed: %v", err)
	}
	if err := tx.QueryRow(`INSERT INTO films (title) VALUES ('A') RETURNING id`).Scan(&filmID); err != nil {
		t.Fatalf("insert film failed: %v", err)
	}

	if _, err := tx.Exec(`
		INSERT INTO movie_proposals
		(title, description, small_description, duration, release_date, country, poster, source, status, proposed_by_user_id)
		VALUES ('M1', 'D', 'S', 120, 2000, 'US', 'poster.jpg', 'manual', 'pending', $1)
	`, proposerID); err != nil {
		t.Fatalf("insert pending proposal failed: %v", err)
	}

	if _, err := tx.Exec(`
		INSERT INTO movie_proposals
		(title, description, small_description, duration, release_date, country, poster, source, status, proposed_by_user_id, moderated_by_user_id, moderated_at)
		VALUES ('M2', 'D', 'S', 120, 2000, 'US', 'poster.jpg', 'manual', 'approved', $1, $2, NOW())
	`, proposerID, moderatorID); err == nil {
		t.Fatal("expected approved proposal without film_id to violate audit check")
	}

	if _, err := tx.Exec(`
		INSERT INTO movie_proposals
		(title, description, small_description, duration, release_date, country, poster, source, status, proposed_by_user_id, moderated_by_user_id, moderated_at, film_id)
		VALUES ('M3', 'D', 'S', 120, 2000, 'US', 'poster.jpg', 'manual', 'approved', $1, $2, NOW(), $3)
	`, proposerID, moderatorID, filmID); err != nil {
		t.Fatalf("insert approved proposal failed: %v", err)
	}

	if _, err := tx.Exec(`DELETE FROM users WHERE id = $1`, moderatorID); err == nil {
		t.Fatal("expected delete moderator to fail due RESTRICT FK")
	}
	if _, err := tx.Exec(`DELETE FROM films WHERE id = $1`, filmID); err == nil {
		t.Fatal("expected delete approved film to fail due RESTRICT FK")
	}
}
