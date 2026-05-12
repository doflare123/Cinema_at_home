package services

import (
	"cinema/internal/api"
	appErrors "cinema/internal/errors"
	"cinema/internal/logger"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type FilmService interface {
	Create(title string) error
	List() ([]dto.FilmView, error)
	GetByID(id uint) (dto.FilmView, error)
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

func (s *filmService) GetFilmsOnVoteSelection() (dto.Film, error) {

	return dto.Film{}, nil
}

func (s *filmService) List() ([]dto.FilmView, error) {
	var films []models.Film
	if err := s.rep.Model(&models.Film{}).Order("title ASC").Find(&films).Error; err != nil {
		return nil, err
	}

	items := make([]dto.FilmView, 0, len(films))
	for _, film := range films {
		items = append(items, mapFilmView(film))
	}

	return items, nil
}

func (s *filmService) GetByID(id uint) (dto.FilmView, error) {
	var film models.Film
	if err := s.rep.First(&film, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.FilmView{}, appErrors.ErrFilmNotFound
		}
		return dto.FilmView{}, err
	}

	return mapFilmView(film), nil
}

func mapFilmView(film models.Film) dto.FilmView {
	return dto.FilmView{
		ID:          film.ID,
		Title:       film.Title,
		Description: film.Description,
		Poster:      film.Poster,
		ReleaseDate: film.ReleaseDate,
	}
}
