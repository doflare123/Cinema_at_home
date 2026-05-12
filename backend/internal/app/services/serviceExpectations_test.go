package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestExpectationServiceMovieSummaryIgnoresRefuseAndPreservesUpsert(t *testing.T) {
	service, _, users, movie, _ := newTestExpectationService(t)

	initialScore := 8
	updatedScore := 9

	vote, err := service.Upsert(users[0].ID, dto.UpsertExpectationRequest{
		TargetType: "movie",
		MovieID:    &movie.ID,
		VoteType:   "score",
		Score:      &initialScore,
		Comment:    "first pass",
	})
	if err != nil {
		t.Fatalf("expected first upsert to succeed, got %v", err)
	}
	if vote.Score == nil || *vote.Score != initialScore {
		t.Fatalf("unexpected first vote: %+v", vote)
	}

	if _, err := service.Upsert(users[0].ID, dto.UpsertExpectationRequest{
		TargetType: "movie",
		MovieID:    &movie.ID,
		VoteType:   "score",
		Score:      &updatedScore,
		Comment:    "updated",
	}); err != nil {
		t.Fatalf("expected second upsert to update, got %v", err)
	}

	if _, err := service.Upsert(users[1].ID, dto.UpsertExpectationRequest{
		TargetType: "movie",
		MovieID:    &movie.ID,
		VoteType:   "refuse",
		Comment:    "skip",
	}); err != nil {
		t.Fatalf("expected refuse upsert to succeed, got %v", err)
	}

	summary, err := service.Summary("movie", movie.ID)
	if err != nil {
		t.Fatalf("expected summary to succeed, got %v", err)
	}

	if !almostEqual(summary.Avg, 9) {
		t.Fatalf("expected avg 9, got %f", summary.Avg)
	}
	if summary.NumericCount != 1 {
		t.Fatalf("expected numeric_count 1, got %d", summary.NumericCount)
	}
	if summary.RefuseCount != 1 {
		t.Fatalf("expected refuse_count 1, got %d", summary.RefuseCount)
	}
	if !summary.ThresholdPassed {
		t.Fatal("expected threshold to pass for avg 9")
	}

	myVotes, err := service.ListByUser(users[0].ID)
	if err != nil {
		t.Fatalf("expected ListByUser to succeed, got %v", err)
	}
	if len(myVotes) != 1 {
		t.Fatalf("expected one user vote, got %d", len(myVotes))
	}
	if myVotes[0].TargetTitle != movie.Title {
		t.Fatalf("expected target title %q, got %q", movie.Title, myVotes[0].TargetTitle)
	}
	if myVotes[0].Score == nil || *myVotes[0].Score != updatedScore {
		t.Fatalf("expected updated score %d, got %+v", updatedScore, myVotes[0].Score)
	}
}

func TestExpectationServiceFranchiseSummaryCountsNumericOnly(t *testing.T) {
	service, _, users, _, franchise := newTestExpectationService(t)

	scoreA := 4
	scoreB := 6

	for i, score := range []int{scoreA, scoreB} {
		if _, err := service.Upsert(users[i].ID, dto.UpsertExpectationRequest{
			TargetType:  "franchise",
			FranchiseID: &franchise.ID,
			VoteType:    "score",
			Score:       &score,
		}); err != nil {
			t.Fatalf("expected score upsert %d to succeed, got %v", i, err)
		}
	}

	if _, err := service.Upsert(users[2].ID, dto.UpsertExpectationRequest{
		TargetType:  "franchise",
		FranchiseID: &franchise.ID,
		VoteType:    "refuse",
	}); err != nil {
		t.Fatalf("expected refuse upsert to succeed, got %v", err)
	}

	summary, err := service.Summary("franchise", franchise.ID)
	if err != nil {
		t.Fatalf("expected summary to succeed, got %v", err)
	}

	if !almostEqual(summary.Avg, 5) {
		t.Fatalf("expected avg 5, got %f", summary.Avg)
	}
	if summary.NumericCount != 2 || summary.RefuseCount != 1 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if !summary.ThresholdPassed {
		t.Fatal("expected threshold to pass for avg 5")
	}
}

func TestExpectationServiceRejectsInvalidPayloads(t *testing.T) {
	service, _, users, movie, franchise := newTestExpectationService(t)

	scoreTooHigh := 11
	_, err := service.Upsert(users[0].ID, dto.UpsertExpectationRequest{
		TargetType: "movie",
		MovieID:    &movie.ID,
		VoteType:   "score",
		Score:      &scoreTooHigh,
	})
	if !errors.Is(err, appErrors.ErrInvalidExpectationScore) {
		t.Fatalf("expected ErrInvalidExpectationScore, got %v", err)
	}

	refuseScore := 3
	_, err = service.Upsert(users[0].ID, dto.UpsertExpectationRequest{
		TargetType:  "franchise",
		FranchiseID: &franchise.ID,
		VoteType:    "refuse",
		Score:       &refuseScore,
	})
	if !errors.Is(err, appErrors.ErrUnexpectedExpectationScore) {
		t.Fatalf("expected ErrUnexpectedExpectationScore, got %v", err)
	}

	_, err = service.Upsert(users[0].ID, dto.UpsertExpectationRequest{
		TargetType:  "movie",
		MovieID:     &movie.ID,
		FranchiseID: &franchise.ID,
		VoteType:    "score",
		Score:       &refuseScore,
	})
	if !errors.Is(err, appErrors.ErrInvalidExpectationTarget) {
		t.Fatalf("expected ErrInvalidExpectationTarget, got %v", err)
	}

	_, err = service.Summary("unknown", movie.ID)
	if !errors.Is(err, appErrors.ErrInvalidExpectationTargetType) {
		t.Fatalf("expected ErrInvalidExpectationTargetType, got %v", err)
	}
}

func TestExpectationServiceRejectsUnknownTarget(t *testing.T) {
	service, _, users, _, _ := newTestExpectationService(t)

	score := 7
	missingMovieID := uint(9999)

	_, err := service.Upsert(users[0].ID, dto.UpsertExpectationRequest{
		TargetType: "movie",
		MovieID:    &missingMovieID,
		VoteType:   "score",
		Score:      &score,
	})
	if !errors.Is(err, appErrors.ErrFilmNotFound) {
		t.Fatalf("expected ErrFilmNotFound, got %v", err)
	}
}

func newTestExpectationService(t *testing.T) (ExpectationService, *testRepository, []*models.User, *models.Film, *models.Franchise) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%d?mode=memory&cache=shared", time.Now().UnixNano())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	rep := &testRepository{db: db}
	for _, model := range []interface{}{
		&models.Role{},
		&models.User{},
		&models.Film{},
		&models.Franchise{},
		&models.FranchiseMovie{},
		&models.ExpectationVote{},
	} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate test db: %v", err)
		}
	}

	for _, role := range []models.Role{{ID: 1, Name: "member"}, {ID: 2, Name: "admin"}} {
		if err := rep.Create(&role).Error; err != nil {
			t.Fatalf("failed to seed role %s: %v", role.Name, err)
		}
	}

	users := make([]*models.User, 0, 3)
	for i := 1; i <= 3; i++ {
		user := &models.User{
			Username:    fmt.Sprintf("user-%d", i),
			DisplayName: fmt.Sprintf("User %d", i),
			Password:    "hashed",
			RoleID:      1,
			Status:      "active",
		}
		if err := rep.Create(user).Error; err != nil {
			t.Fatalf("failed to seed user %d: %v", i, err)
		}
		users = append(users, user)
	}

	movie := &models.Film{
		Title:            "Movie One",
		Description:      "Description",
		SmallDescription: "Short",
		Duration:         100,
		ReleaseDate:      2024,
		Country:          "US",
		Poster:           "poster",
		RatingKp:         7.5,
	}
	if err := rep.Create(movie).Error; err != nil {
		t.Fatalf("failed to seed movie: %v", err)
	}

	franchise := &models.Franchise{
		Title:       "Franchise One",
		Description: "Description",
	}
	if err := rep.Create(franchise).Error; err != nil {
		t.Fatalf("failed to seed franchise: %v", err)
	}

	return NewExpectationService(rep), rep, users, movie, franchise
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}
