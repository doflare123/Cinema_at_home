package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"errors"
	"testing"
)

func TestWeeklyPackServiceVoteLimits(t *testing.T) {
	svc, rep, userID, packID, movieIDs := newTestWeeklyPackService(t)

	_, err := svc.UpsertVote(packID, userID, dto.UpsertWeeklyPackVoteRequest{MovieID: movieIDs[0], Score: intPtr(3)})
	if err != nil {
		t.Fatalf("first +3 vote failed: %v", err)
	}
	_, err = svc.UpsertVote(packID, userID, dto.UpsertWeeklyPackVoteRequest{MovieID: movieIDs[1], Score: intPtr(3)})
	if !errors.Is(err, appErrors.ErrWeeklyPackVoteLimitExceeded) {
		t.Fatalf("expected ErrWeeklyPackVoteLimitExceeded, got %v", err)
	}

	var votes []models.WeeklyPackVote
	if err := rep.Where("pack_id = ? AND user_id = ?", packID, userID).Find(&votes).Error; err != nil {
		t.Fatalf("failed to list votes: %v", err)
	}
	if len(votes) != 1 {
		t.Fatalf("expected 1 persisted vote, got %d", len(votes))
	}
}

func TestWeeklyPackServiceStatusTransitionRequiresMovies(t *testing.T) {
	svc, _, userID, _, _ := newTestWeeklyPackService(t)

	created, err := svc.Create(userID, dto.CreateWeeklyPackRequest{Name: "Empty pack"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	_, err = svc.UpdateStatus(created.ID, dto.UpdateWeeklyPackStatusRequest{Status: "voting"})
	if !errors.Is(err, appErrors.ErrWeeklyPackMustHaveMovies) {
		t.Fatalf("expected ErrWeeklyPackMustHaveMovies, got %v", err)
	}
}

func newTestWeeklyPackService(t *testing.T) (WeeklyPackService, *testRepository, uint, uint, []uint) {
	t.Helper()

	db := openTestSQLiteDB(t)
	rep := &testRepository{db: db}

	for _, model := range []interface{}{
		&models.Role{},
		&models.User{},
		&models.Film{},
		&models.WeeklyPack{},
		&models.WeeklyPackMovie{},
		&models.WeeklyPackVote{},
	} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate model: %v", err)
		}
	}

	if err := rep.Create(&models.Role{ID: 1, Name: "member"}).Error; err != nil {
		t.Fatalf("seed role member: %v", err)
	}
	if err := rep.Create(&models.Role{ID: 2, Name: "admin"}).Error; err != nil {
		t.Fatalf("seed role admin: %v", err)
	}
	user := models.User{Username: "user1", Password: "x", DisplayName: "User One", RoleID: 1, Status: "active"}
	if err := rep.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	admin := models.User{Username: "admin1", Password: "x", DisplayName: "Admin One", RoleID: 2, Status: "active"}
	if err := rep.Create(&admin).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	f1 := models.Film{Title: "WP Alpha", Description: "d", SmallDescription: "s", Duration: 100, ReleaseDate: 2000, Country: "US", Poster: "a.jpg", RatingKp: 7.1}
	f2 := models.Film{Title: "WP Beta", Description: "d", SmallDescription: "s", Duration: 100, ReleaseDate: 2001, Country: "US", Poster: "b.jpg", RatingKp: 7.2}
	if err := rep.Create(&f1).Error; err != nil {
		t.Fatalf("seed film1: %v", err)
	}
	if err := rep.Create(&f2).Error; err != nil {
		t.Fatalf("seed film2: %v", err)
	}

	pack := models.WeeklyPack{Name: "Week Pack", Status: weeklyPackStatusVoting, CreatedByUserID: admin.ID}
	if err := rep.Create(&pack).Error; err != nil {
		t.Fatalf("seed pack: %v", err)
	}
	if err := rep.Create(&models.WeeklyPackMovie{PackID: pack.ID, MovieID: f1.ID, SortOrder: 1}).Error; err != nil {
		t.Fatalf("seed pack movie1: %v", err)
	}
	if err := rep.Create(&models.WeeklyPackMovie{PackID: pack.ID, MovieID: f2.ID, SortOrder: 2}).Error; err != nil {
		t.Fatalf("seed pack movie2: %v", err)
	}

	return NewWeeklyPackService(rep), rep, user.ID, pack.ID, []uint{f1.ID, f2.ID}
}

func intPtr(v int) *int { return &v }
