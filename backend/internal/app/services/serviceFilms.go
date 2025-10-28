package services

import (
	"cinema/internal/container"
)

type FilmsService interface {
}

type filmsService struct {
	container *container.Container
}

func NewFilmsService(container *container.Container) FilmsService {
	return &filmsService{
		container: container,
	}
}

// func (s *filmsService) CreateFilm(dto dto.AboutFilm) (*models.Film, error) {
// 	film := &models.Film{
// 		Title:            dto.Title,
// 		Description:      dto.Description,
// 		ShortDescription: dto.ShortDescription,
// 		Duration:         int32(dto.Duration),
// 		ReleaseDate:      dto.ReleaseDate,
// 		Country:          dto.Country,
// 		Poster:           dto.Poster,
// 		RatinKp:          float64(dto.RatingKp),
// 	}
// 	err := s.container.DB.Create(film).Error
// 	if err != nil {
// 		return nil, err
// 	}
// 	return film, nil
// }
