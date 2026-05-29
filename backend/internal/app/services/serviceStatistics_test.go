package services

import (
	"cinema/internal/models"
	"testing"
)

func TestStatisticsServiceSummaryUsesRawData(t *testing.T) {
	db := openTestSQLiteDB(t)
	rep := &testRepository{db: db}

	for _, model := range []interface{}{
		&models.Role{},
		&models.User{},
		&models.Film{},
		&models.Franchise{},
		&models.MovieProposal{},
		&models.ExpectationVote{},
		&models.Review{},
		&models.WeeklyPack{},
		&models.WeeklyPackMovie{},
		&models.WeeklyPackVote{},
	} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate model: %v", err)
		}
	}

	if err := rep.Create(&models.Role{ID: 1, Name: "member"}).Error; err != nil {
		t.Fatalf("seed role failed: %v", err)
	}
	active := models.User{Username: "active", Password: "x", DisplayName: "Active", RoleID: 1, Status: "active"}
	pending := models.User{Username: "pending", Password: "x", DisplayName: "Pending", RoleID: 1, Status: "pending"}
	if err := rep.Create(&active).Error; err != nil {
		t.Fatalf("seed active user failed: %v", err)
	}
	if err := rep.Create(&pending).Error; err != nil {
		t.Fatalf("seed pending user failed: %v", err)
	}

	film1 := models.Film{Title: "A", Description: "A", SmallDescription: "A", Duration: 100, ReleaseDate: 2000, Country: "US", Poster: "a.jpg", RatingKp: 7}
	film2 := models.Film{Title: "B", Description: "B", SmallDescription: "B", Duration: 110, ReleaseDate: 2001, Country: "US", Poster: "b.jpg", RatingKp: 8}
	if err := rep.Create(&film1).Error; err != nil {
		t.Fatalf("seed film1 failed: %v", err)
	}
	if err := rep.Create(&film2).Error; err != nil {
		t.Fatalf("seed film2 failed: %v", err)
	}

	franchise := models.Franchise{Title: "F1", Description: "Desc"}
	if err := rep.Create(&franchise).Error; err != nil {
		t.Fatalf("seed franchise failed: %v", err)
	}

	approved := models.MovieProposal{Title: "P1", Description: "D", SmallDescription: "S", Duration: 90, ReleaseDate: 1999, Country: "US", Poster: "p.jpg", RatingKp: 7, Source: "manual", Status: "approved", ProposedByUserID: active.ID, ModeratedByUserID: &active.ID, FilmID: &film1.ID}
	rejected := models.MovieProposal{Title: "P2", Description: "D", SmallDescription: "S", Duration: 90, ReleaseDate: 1999, Country: "US", Poster: "p.jpg", RatingKp: 7, Source: "manual", Status: "rejected", ProposedByUserID: active.ID, ModeratedByUserID: &active.ID}
	pendingProposal := models.MovieProposal{Title: "P3", Description: "D", SmallDescription: "S", Duration: 90, ReleaseDate: 1999, Country: "US", Poster: "p.jpg", RatingKp: 7, Source: "manual", Status: "pending", ProposedByUserID: active.ID}
	if err := rep.Create(&approved).Error; err != nil {
		t.Fatalf("seed approved proposal failed: %v", err)
	}
	if err := rep.Create(&rejected).Error; err != nil {
		t.Fatalf("seed rejected proposal failed: %v", err)
	}
	if err := rep.Create(&pendingProposal).Error; err != nil {
		t.Fatalf("seed pending proposal failed: %v", err)
	}

	score := 8
	if err := rep.Create(&models.ExpectationVote{UserID: active.ID, TargetType: "movie", MovieID: &film1.ID, VoteType: "score", Score: &score}).Error; err != nil {
		t.Fatalf("seed score expectation failed: %v", err)
	}
	if err := rep.Create(&models.ExpectationVote{UserID: pending.ID, TargetType: "movie", MovieID: &film1.ID, VoteType: "refuse"}).Error; err != nil {
		t.Fatalf("seed refuse expectation failed: %v", err)
	}

	if err := rep.Create(&models.Review{FilmID: film1.ID, UserID: active.ID, Mode: "simple", Score: &score, FinalScore: 8, CriteriaScoresJSON: "{}"}).Error; err != nil {
		t.Fatalf("seed review failed: %v", err)
	}

	pack := models.WeeklyPack{Name: "Pack", Status: "voting", CreatedByUserID: active.ID}
	if err := rep.Create(&pack).Error; err != nil {
		t.Fatalf("seed pack failed: %v", err)
	}
	if err := rep.Create(&models.WeeklyPackMovie{PackID: pack.ID, MovieID: film1.ID, SortOrder: 1}).Error; err != nil {
		t.Fatalf("seed pack movie failed: %v", err)
	}
	if err := rep.Create(&models.WeeklyPackVote{PackID: pack.ID, MovieID: film1.ID, UserID: active.ID, Score: 3}).Error; err != nil {
		t.Fatalf("seed pack vote failed: %v", err)
	}

	svc := NewStatisticsService(rep)
	summary, err := svc.Summary()
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}

	if summary.MoviesTotal != 2 || summary.FranchisesTotal != 1 || summary.UsersActive != 1 {
		t.Fatalf("unexpected base counters: %+v", summary)
	}
	if summary.ProposalsPending != 1 || summary.ProposalsApproved != 1 || summary.ProposalsRejected != 1 {
		t.Fatalf("unexpected proposal counters: %+v", summary)
	}
	if summary.ExpectationsTotal != 2 || summary.ExpectationsNumeric != 1 || summary.ExpectationsRefuse != 1 || summary.ExpectationsAverage != 8 {
		t.Fatalf("unexpected expectation counters: %+v", summary)
	}
	if summary.ReviewsTotal != 1 || summary.ReviewsAverage != 8 {
		t.Fatalf("unexpected review counters: %+v", summary)
	}
	if summary.WeeklyPacksTotal != 1 || summary.WeeklyPacksVoting != 1 || summary.WeeklyPackVotesTotal != 1 || summary.WeeklyPackMoviesTotal != 1 {
		t.Fatalf("unexpected weekly pack counters: %+v", summary)
	}
}
