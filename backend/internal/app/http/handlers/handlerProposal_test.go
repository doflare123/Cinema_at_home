package handlers

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type stubMovieProposalService struct {
	create     func(userID uint, req dto.CreateMovieProposalRequest) (dto.MovieProposalView, error)
	list       func(status string) ([]dto.MovieProposalView, error)
	listByUser func(userID uint) ([]dto.MovieProposalView, error)
	getByID    func(id uint) (dto.MovieProposalView, error)
	moderate   func(proposalID, adminID uint, req dto.ModerateMovieProposalRequest) (dto.MovieProposalView, error)
}

func (s stubMovieProposalService) Create(userID uint, req dto.CreateMovieProposalRequest) (dto.MovieProposalView, error) {
	if s.create != nil {
		return s.create(userID, req)
	}
	return dto.MovieProposalView{}, nil
}

func (s stubMovieProposalService) List(status string) ([]dto.MovieProposalView, error) {
	if s.list != nil {
		return s.list(status)
	}
	return nil, nil
}

func (s stubMovieProposalService) ListByUser(userID uint) ([]dto.MovieProposalView, error) {
	if s.listByUser != nil {
		return s.listByUser(userID)
	}
	return nil, nil
}

func (s stubMovieProposalService) GetByID(id uint) (dto.MovieProposalView, error) {
	if s.getByID != nil {
		return s.getByID(id)
	}
	return dto.MovieProposalView{}, nil
}

func (s stubMovieProposalService) Moderate(proposalID, adminID uint, req dto.ModerateMovieProposalRequest) (dto.MovieProposalView, error) {
	if s.moderate != nil {
		return s.moderate(proposalID, adminID, req)
	}
	return dto.MovieProposalView{}, nil
}

func TestMovieProposalHandlerCreateReturnsCreated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewMovieProposalHandler(stubMovieProposalService{
		create: func(userID uint, req dto.CreateMovieProposalRequest) (dto.MovieProposalView, error) {
			return dto.MovieProposalView{ID: 10, Title: req.Title, Status: "pending", ProposedByUserID: userID}, nil
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Set("user_id", uint(7))
	ctx.Request = httptest.NewRequest(http.MethodPost, "/proposals", strings.NewReader(`{
		"title":"Alpha",
		"description":"Long",
		"small_description":"Short",
		"duration":100,
		"release_date":2000,
		"country":"US",
		"poster":"poster.jpg",
		"rating_kp":7.1
	}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Create(ctx)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}
}

func TestMovieProposalHandlerModerateMapsDuplicateFilm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewMovieProposalHandler(stubMovieProposalService{
		moderate: func(proposalID, adminID uint, req dto.ModerateMovieProposalRequest) (dto.MovieProposalView, error) {
			return dto.MovieProposalView{}, appErrors.ErrMovieProposalDuplicateFilm
		},
	})

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "42"}}
	ctx.Set("user_id", uint(2))
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/admin/proposals/42/status", strings.NewReader(`{"status":"approved"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.Moderate(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}
