package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"errors"
	"testing"
)

func TestFilmServiceListReturnsFrontendContract(t *testing.T) {
	service, _ := newTestFilmService(t)

	items, err := service.List()
	if err != nil {
		t.Fatalf("expected List to succeed, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	first := items[0]
	if first.ID == 0 || first.Title == "" || first.Description == "" || first.Poster == "" || first.ReleaseDate == 0 {
		t.Fatalf("expected populated frontend film contract, got %+v", first)
	}
	if first.Title != "Alpha" {
		t.Fatalf("expected items sorted by title, got first title %q", first.Title)
	}
}

func TestFilmServiceGetByIDReturnsFrontendContract(t *testing.T) {
	service, film := newTestFilmService(t)

	item, err := service.GetByID(film.ID)
	if err != nil {
		t.Fatalf("expected GetByID to succeed, got %v", err)
	}

	if item.ID != film.ID || item.Title != film.Title || item.Description != film.Description || item.Poster != film.Poster || item.ReleaseDate != film.ReleaseDate {
		t.Fatalf("unexpected film view: %+v", item)
	}
}

func TestFilmServiceGetByIDReturnsNotFound(t *testing.T) {
	service, _ := newTestFilmService(t)

	_, err := service.GetByID(9999)
	if !errors.Is(err, appErrors.ErrFilmNotFound) {
		t.Fatalf("expected ErrFilmNotFound, got %v", err)
	}
}

func newTestFilmService(t *testing.T) (FilmService, *models.Film) {
	t.Helper()

	db := openTestSQLiteDB(t)

	rep := &testRepository{db: db}
	if err := rep.AutoMigrate(&models.Film{}); err != nil {
		t.Fatalf("failed to migrate film model: %v", err)
	}

	alpha := &models.Film{
		Title:            "Alpha",
		Description:      "Alpha description",
		SmallDescription: "Alpha short",
		Duration:         100,
		ReleaseDate:      2001,
		Country:          "US",
		Poster:           "alpha.jpg",
		RatingKp:         7.1,
	}
	if err := rep.Create(alpha).Error; err != nil {
		t.Fatalf("failed to seed alpha film: %v", err)
	}

	bravo := &models.Film{
		Title:            "Bravo",
		Description:      "Bravo description",
		SmallDescription: "Bravo short",
		Duration:         110,
		ReleaseDate:      2002,
		Country:          "UK",
		Poster:           "bravo.jpg",
		RatingKp:         7.3,
	}
	if err := rep.Create(bravo).Error; err != nil {
		t.Fatalf("failed to seed bravo film: %v", err)
	}

	return NewFilmService(noopLogger{}, rep, ""), alpha
}
