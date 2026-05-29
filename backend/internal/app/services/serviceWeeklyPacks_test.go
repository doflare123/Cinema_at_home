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

func TestWeeklyPackServiceGetCurrentVoting(t *testing.T) {
	svc, _, _, packID, _ := newTestWeeklyPackService(t)

	current, err := svc.GetCurrentVoting()
	if err != nil {
		t.Fatalf("GetCurrentVoting failed: %v", err)
	}
	if current.ID != packID {
		t.Fatalf("expected current pack %d, got %d", packID, current.ID)
	}
}

func TestWeeklyPackServiceGetCurrentVotingNotFound(t *testing.T) {
	svc, rep, _, packID, _ := newTestWeeklyPackService(t)

	if err := rep.Model(&models.WeeklyPack{}).Where("id = ?", packID).Update("status", "closed").Error; err != nil {
		t.Fatalf("close pack failed: %v", err)
	}

	_, err := svc.GetCurrentVoting()
	if !errors.Is(err, appErrors.ErrWeeklyPackCurrentNotFound) {
		t.Fatalf("expected ErrWeeklyPackCurrentNotFound, got %v", err)
	}
}

func TestWeeklyPackServiceGetUserVoteLimits(t *testing.T) {
	svc, _, userID, packID, movieIDs := newTestWeeklyPackService(t)

	if _, err := svc.UpsertVote(packID, userID, dto.UpsertWeeklyPackVoteRequest{MovieID: movieIDs[0], Score: intPtr(3)}); err != nil {
		t.Fatalf("vote +3 failed: %v", err)
	}
	if _, err := svc.UpsertVote(packID, userID, dto.UpsertWeeklyPackVoteRequest{MovieID: movieIDs[1], Score: intPtr(0)}); err != nil {
		t.Fatalf("vote 0 failed: %v", err)
	}

	limits, err := svc.GetUserVoteLimits(packID, userID)
	if err != nil {
		t.Fatalf("GetUserVoteLimits failed: %v", err)
	}
	if limits.PackID != packID {
		t.Fatalf("expected pack id %d, got %d", packID, limits.PackID)
	}
	if !limits.ZeroScoreUnlimited {
		t.Fatalf("expected ZeroScoreUnlimited=true")
	}
	if limits.ZeroScoreUsed != 1 {
		t.Fatalf("expected ZeroScoreUsed=1, got %d", limits.ZeroScoreUsed)
	}

	if len(limits.Limits) != 4 {
		t.Fatalf("expected 4 limited scores, got %d", len(limits.Limits))
	}
	first := limits.Limits[0]
	if first.Score != 3 || first.Limit != 1 || first.Used != 1 || first.Remaining != 0 {
		t.Fatalf("unexpected +3 limit view: %+v", first)
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
