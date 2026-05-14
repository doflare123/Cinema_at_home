package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"errors"
	"testing"
)

func TestMovieProposalServiceCreatesPendingProposal(t *testing.T) {
	svc, _, memberID, _ := newTestMovieProposalService(t)

	proposal, err := svc.Create(memberID, validMovieProposalRequest("Alpha"))
	if err != nil {
		t.Fatalf("create proposal failed: %v", err)
	}

	if proposal.Status != "pending" || proposal.ProposedByUserID != memberID {
		t.Fatalf("unexpected proposal audit fields: %+v", proposal)
	}
	if proposal.FilmID != nil || proposal.ModeratedAt != nil || proposal.ModeratedByUserID != nil {
		t.Fatalf("new proposal must not be moderated: %+v", proposal)
	}
}

func TestMovieProposalServiceApproveCreatesFilmAndAudit(t *testing.T) {
	svc, rep, memberID, adminID := newTestMovieProposalService(t)

	proposal, err := svc.Create(memberID, validMovieProposalRequest("Approve Me"))
	if err != nil {
		t.Fatalf("create proposal failed: %v", err)
	}

	approved, err := svc.Moderate(proposal.ID, adminID, dto.ModerateMovieProposalRequest{
		Status:            "approved",
		ModerationComment: "ok",
	})
	if err != nil {
		t.Fatalf("approve proposal failed: %v", err)
	}

	if approved.Status != "approved" || approved.FilmID == nil || approved.ModeratedByUserID == nil || *approved.ModeratedByUserID != adminID {
		t.Fatalf("unexpected approved proposal: %+v", approved)
	}
	if approved.ModeratedAt == nil || approved.ModerationComment != "ok" {
		t.Fatalf("expected moderation audit to be filled: %+v", approved)
	}

	var film models.Film
	if err := rep.First(&film, *approved.FilmID).Error; err != nil {
		t.Fatalf("expected created film: %v", err)
	}
	if film.Title != "Approve Me" {
		t.Fatalf("unexpected created film: %+v", film)
	}
}

func TestMovieProposalServiceRejectsClosedProposalModeration(t *testing.T) {
	svc, _, memberID, adminID := newTestMovieProposalService(t)

	proposal, err := svc.Create(memberID, validMovieProposalRequest("Reject Me"))
	if err != nil {
		t.Fatalf("create proposal failed: %v", err)
	}
	if _, err := svc.Moderate(proposal.ID, adminID, dto.ModerateMovieProposalRequest{Status: "rejected"}); err != nil {
		t.Fatalf("reject proposal failed: %v", err)
	}

	_, err = svc.Moderate(proposal.ID, adminID, dto.ModerateMovieProposalRequest{Status: "approved"})
	if !errors.Is(err, appErrors.ErrMovieProposalAlreadyClosed) {
		t.Fatalf("expected ErrMovieProposalAlreadyClosed, got %v", err)
	}
}

func TestMovieProposalServiceApproveLinksExistingFilm(t *testing.T) {
	svc, rep, memberID, adminID := newTestMovieProposalService(t)
	existing := models.Film{
		Title:            "Duplicate",
		Description:      "Existing",
		SmallDescription: "Existing",
		Duration:         90,
		ReleaseDate:      2020,
		Country:          "US",
		Poster:           "poster.jpg",
		RatingKp:         7,
	}
	if err := rep.Create(&existing).Error; err != nil {
		t.Fatalf("seed film failed: %v", err)
	}

	proposal, err := svc.Create(memberID, validMovieProposalRequest("duplicate"))
	if err != nil {
		t.Fatalf("create proposal failed: %v", err)
	}

	approved, err := svc.Moderate(proposal.ID, adminID, dto.ModerateMovieProposalRequest{Status: "approved"})
	if err != nil {
		t.Fatalf("approve proposal failed: %v", err)
	}
	if approved.FilmID == nil || *approved.FilmID != existing.ID {
		t.Fatalf("expected proposal to link existing film %d, got %+v", existing.ID, approved)
	}
}

func TestMovieProposalServiceApproveRejectsAmbiguousExistingFilm(t *testing.T) {
	svc, rep, memberID, adminID := newTestMovieProposalService(t)
	if err := rep.Create(&models.Film{
		Title:            "Ambiguous",
		Description:      "Existing",
		SmallDescription: "Existing",
		Duration:         90,
		ReleaseDate:      1999,
		Country:          "US",
		Poster:           "poster.jpg",
		RatingKp:         7,
	}).Error; err != nil {
		t.Fatalf("seed film failed: %v", err)
	}

	proposal, err := svc.Create(memberID, validMovieProposalRequest("ambiguous"))
	if err != nil {
		t.Fatalf("create proposal failed: %v", err)
	}

	_, err = svc.Moderate(proposal.ID, adminID, dto.ModerateMovieProposalRequest{Status: "approved"})
	if !errors.Is(err, appErrors.ErrMovieProposalDuplicateFilm) {
		t.Fatalf("expected ErrMovieProposalDuplicateFilm, got %v", err)
	}
}

func TestMovieProposalServiceRejectsInvalidSource(t *testing.T) {
	svc, _, memberID, _ := newTestMovieProposalService(t)
	req := validMovieProposalRequest("Bad Source")
	req.Source = "unknown"

	_, err := svc.Create(memberID, req)
	if !errors.Is(err, appErrors.ErrInvalidMovieProposalPayload) {
		t.Fatalf("expected ErrInvalidMovieProposalPayload, got %v", err)
	}
}

func TestMovieProposalServiceAcceptsKnownSources(t *testing.T) {
	for _, tc := range []struct {
		name     string
		source   string
		expected string
	}{
		{name: "empty defaults to manual", source: "", expected: "manual"},
		{name: "manual", source: "manual", expected: "manual"},
		{name: "kinopoisk", source: "kinopoisk", expected: "kinopoisk"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, _, memberID, _ := newTestMovieProposalService(t)
			req := validMovieProposalRequest("Source " + tc.name)
			req.Source = tc.source

			proposal, err := svc.Create(memberID, req)
			if err != nil {
				t.Fatalf("create proposal failed: %v", err)
			}
			if proposal.Source != tc.expected {
				t.Fatalf("expected source %q, got %q", tc.expected, proposal.Source)
			}
		})
	}
}

func newTestMovieProposalService(t *testing.T) (MovieProposalService, *testRepository, uint, uint) {
	t.Helper()

	db := openTestSQLiteDB(t)
	rep := &testRepository{db: db}
	for _, model := range []interface{}{
		&models.Role{},
		&models.User{},
		&models.Film{},
		&models.MovieProposal{},
	} {
		if err := rep.AutoMigrate(model); err != nil {
			t.Fatalf("failed to migrate model: %v", err)
		}
	}

	if err := rep.Create(&models.Role{ID: 1, Name: "member"}).Error; err != nil {
		t.Fatalf("seed member role: %v", err)
	}
	if err := rep.Create(&models.Role{ID: 2, Name: "admin"}).Error; err != nil {
		t.Fatalf("seed admin role: %v", err)
	}

	member := models.User{Username: "member", Password: "x", DisplayName: "Member", RoleID: 1, Status: "active"}
	if err := rep.Create(&member).Error; err != nil {
		t.Fatalf("seed member: %v", err)
	}
	admin := models.User{Username: "admin", Password: "x", DisplayName: "Admin", RoleID: 2, Status: "active"}
	if err := rep.Create(&admin).Error; err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	return NewMovieProposalService(rep), rep, member.ID, admin.ID
}

func validMovieProposalRequest(title string) dto.CreateMovieProposalRequest {
	return dto.CreateMovieProposalRequest{
		Title:            title,
		Description:      "Long description",
		SmallDescription: "Short description",
		Duration:         120,
		ReleaseDate:      2020,
		Country:          "US",
		Poster:           "poster.jpg",
		RatingKp:         7.5,
	}
}
