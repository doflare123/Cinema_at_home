package services

import "cinema/internal/container"

type FilmsService interface {
	GetAllFilms()
}

type filmsService struct {
	container *container.Container
}

func NewFilmsService(container *container.Container) FilmsService {
	return &filmsService{
		container: container,
	}
}

func (s *filmsService) GetAllFilms()
