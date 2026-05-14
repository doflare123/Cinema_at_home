package services

import (
	"cinema/internal/api"
	appErrors "cinema/internal/errors"
	"errors"
	"testing"
)

type fakeKinopoiskSearcher struct {
	results []api.FilmResult
	err     error
}

func (f fakeKinopoiskSearcher) SearchFilms(query string, limit int) ([]api.FilmResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.results, nil
}

func TestKinopoiskServiceSearchMapsResults(t *testing.T) {
	svc := NewKinopoiskService(fakeKinopoiskSearcher{
		results: []api.FilmResult{{
			SourceID:         "42",
			Title:            "Movie",
			Description:      "Long",
			ShortDescription: "Short",
			Duration:         100,
			Year:             2000,
			Countries:        []string{"US", "UK"},
			Genres:           []string{"drama"},
			Poster:           "poster.jpg",
			RatingKp:         7.2,
		}},
	})

	results, err := svc.Search("movie", 1)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].ProviderMovieID != "42" || results[0].Source != "kinopoisk" || results[0].Country != "US, UK" || results[0].Duration != 100 {
		t.Fatalf("unexpected mapped result: %+v", results[0])
	}
}

func TestKinopoiskServiceSearchRejectsEmptyQuery(t *testing.T) {
	svc := NewKinopoiskService(fakeKinopoiskSearcher{})

	_, err := svc.Search(" ", 1)
	if !errors.Is(err, appErrors.ErrKinopoiskQueryRequired) {
		t.Fatalf("expected ErrKinopoiskQueryRequired, got %v", err)
	}
}

func TestKinopoiskServiceSearchMapsProviderError(t *testing.T) {
	upstreamErr := errors.New("upstream")
	svc := NewKinopoiskService(fakeKinopoiskSearcher{err: upstreamErr})

	_, err := svc.Search("movie", 1)
	if !errors.Is(err, appErrors.ErrKinopoiskSearchFailed) {
		t.Fatalf("expected ErrKinopoiskSearchFailed, got %v", err)
	}
	if !errors.Is(err, upstreamErr) {
		t.Fatalf("expected upstream error to remain in chain, got %v", err)
	}
}
