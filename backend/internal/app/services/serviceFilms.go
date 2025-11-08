package services

import (
	"cinema/internal/api"
	"cinema/internal/logger"
	"cinema/internal/models"
	"cinema/internal/repository"
	"fmt"
	"strings"
)

type FilmService interface {
	Create(title string) error
}

type filmService struct {
	logger logger.Logger
	rep    repository.Repository
	apiKey string
}

func NewFilmService(logger logger.Logger, rep repository.Repository, apiKey string) FilmService {
	return &filmService{
		logger: logger,
		rep:    rep,
		apiKey: apiKey,
	}
}

func (s *filmService) Create(title string) error {
	var film models.Film

	if err := film.NameAlreadyExist(s.rep, title); err == nil {
		return fmt.Errorf("Фильм с именем '%s' уже есть", title)
	}

	raw, err := api.SearchFilm(title, s.apiKey)
	if err != nil {
		s.logger.Error("Ошибка при поиске фильма: ", err)
		return err
	}

	return s.rep.Transaction(func(tx repository.Repository) error {

		film := models.Film{
			Title:            raw.Title,
			Description:      raw.Description,
			SmallDescription: raw.ShortDescription,
			Duration:         int32(raw.Duration),
			ReleaseDate:      raw.Year,
			Country:          strings.Join(raw.Countries, ", "),
			Poster:           raw.Poster,
			RatingKp:         raw.RatingKp,
		}

		if err := film.Create(tx); err != nil {
			return err
		}

		for _, g := range raw.Genres {
			var genre models.Genre

			err := genre.NameAlreadyExist(tx, g)
			if err != nil {
				genre = models.Genre{Name: g}
				if err := genre.Create(tx); err != nil {
					return fmt.Errorf("failed to create genre '%s': %w", g, err)
				}
			}

			fg := models.FilmGenre{
				FilmID:  film.ID,
				GenreID: genre.ID,
			}

			if err := tx.Create(&fg).Error; err != nil {
				return fmt.Errorf("failed to create filmGenre: %w", err)
			}
		}

		return nil
	})
}
