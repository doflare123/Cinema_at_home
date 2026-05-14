package services

import (
	"cinema/internal/api"
	appErrors "cinema/internal/errors"
	"cinema/internal/models/dto"
	"fmt"
	"strings"
)

type KinopoiskSearcher interface {
	SearchFilms(query string, limit int) ([]api.FilmResult, error)
}

type KinopoiskService interface {
	Search(query string, limit int) ([]dto.KinopoiskSearchResult, error)
}

type kinopoiskService struct {
	searcher KinopoiskSearcher
}

func NewKinopoiskService(searcher KinopoiskSearcher) KinopoiskService {
	return &kinopoiskService{searcher: searcher}
}

func (s *kinopoiskService) Search(query string, limit int) ([]dto.KinopoiskSearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, appErrors.ErrKinopoiskQueryRequired
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	results, err := s.searcher.SearchFilms(query, limit)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", appErrors.ErrKinopoiskSearchFailed, err)
	}

	out := make([]dto.KinopoiskSearchResult, 0, len(results))
	for _, result := range results {
		out = append(out, dto.KinopoiskSearchResult{
			ProviderMovieID:  result.SourceID,
			Title:            result.Title,
			Description:      result.Description,
			SmallDescription: result.ShortDescription,
			Duration:         int32(result.Duration),
			ReleaseDate:      result.Year,
			Country:          strings.Join(result.Countries, ", "),
			Poster:           result.Poster,
			RatingKp:         result.RatingKp,
			Genres:           result.Genres,
			Source:           "kinopoisk",
		})
	}

	return out, nil
}
