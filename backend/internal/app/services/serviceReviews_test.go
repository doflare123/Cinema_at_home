package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"errors"
	"testing"
	"time"
)

func TestReviewServiceUpsertsSimpleReview(t *testing.T) {
	svc, _, userID, movieID := newTestReviewService(t)
	score := 8

	review, err := svc.Upsert(movieID, userID, dto.UpsertReviewRequest{
		Mode:    "simple",
		Score:   &score,
		Comment: "solid",
	})
	if err != nil {
		t.Fatalf("upsert simple review failed: %v", err)
	}
	if review.FinalScore != 8 || review.Score == nil || *review.Score != 8 {
		t.Fatalf("unexpected simple review score: %+v", review)
	}
	firstUpdatedAt := review.UpdatedAt

	nextScore := 6
	time.Sleep(10 * time.Millisecond)
	review, err = svc.Upsert(movieID, userID, dto.UpsertReviewRequest{
		Mode:  "simple",
		Score: &nextScore,
	})
	if err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}
	if review.FinalScore != 6 {
		t.Fatalf("expected updated final score 6, got %v", review.FinalScore)
	}
	if !review.UpdatedAt.After(firstUpdatedAt) {
		t.Fatalf("expected updated_at to advance after upsert, before=%s after=%s", firstUpdatedAt, review.UpdatedAt)
	}
}

func TestReviewServiceCriteriaReviewCalculatesAverage(t *testing.T) {
	svc, _, userID, movieID := newTestReviewService(t)

	review, err := svc.Upsert(movieID, userID, dto.UpsertReviewRequest{
		Mode: "criteria",
		CriteriaScores: map[string]int{
			"plot":   8,
			"visual": 10,
		},
	})
	if err != nil {
		t.Fatalf("upsert criteria review failed: %v", err)
	}
	if review.FinalScore != 9 {
		t.Fatalf("expected average 9, got %v", review.FinalScore)
	}
	if len(review.CriteriaScores) != 2 {
		t.Fatalf("expected 2 criteria scores, got %d", len(review.CriteriaScores))
	}
}

func TestReviewServiceRejectsManualScoreForCriteriaMode(t *testing.T) {
	svc, _, userID, movieID := newTestReviewService(t)
	score := 7

	_, err := svc.Upsert(movieID, userID, dto.UpsertReviewRequest{
		Mode:  "criteria",
		Score: &score,
		CriteriaScores: map[string]int{
			"plot": 8,
		},
	})
	if !errors.Is(err, appErrors.ErrUnexpectedReviewScore) {
		t.Fatalf("expected ErrUnexpectedReviewScore, got %v", err)
	}
}

func TestReviewServiceRejectsManualFinalScore(t *testing.T) {
	svc, _, userID, movieID := newTestReviewService(t)
	score := 7
	finalScore := 10.0

	_, err := svc.Upsert(movieID, userID, dto.UpsertReviewRequest{
		Mode:       "simple",
		Score:      &score,
		FinalScore: &finalScore,
	})
	if !errors.Is(err, appErrors.ErrManualReviewFinalScoreEdit) {
		t.Fatalf("expected ErrManualReviewFinalScoreEdit, got %v", err)
	}
}

func newTestReviewService(t *testing.T) (ReviewService, *testRepository, uint, uint) {
	t.Helper()

	db := openTestSQLiteDB(t)
	rep := &testRepository{db: db}

	for _, model := range []interface{}{
		&models.Role{},
		&models.User{},
		&models.Film{},
		&models.Review{},
	} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate model: %v", err)
		}
	}

	if err := rep.Create(&models.Role{ID: 1, Name: "member"}).Error; err != nil {
		t.Fatalf("seed role member: %v", err)
	}
	user := models.User{Username: "reviewer", Password: "x", DisplayName: "Reviewer", RoleID: 1, Status: "active"}
	if err := rep.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	film := models.Film{Title: "Review Film", Description: "d", SmallDescription: "s", Duration: 100, ReleaseDate: 2000, Country: "US", Poster: "poster.jpg", RatingKp: 7.1}
	if err := rep.Create(&film).Error; err != nil {
		t.Fatalf("seed film: %v", err)
	}

	return NewReviewService(rep), rep, user.ID, film.ID
}
