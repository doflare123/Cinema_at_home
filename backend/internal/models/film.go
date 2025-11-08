package models

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/repository"
)

type Film struct {
	ID               uint    `gorm:"primaryKey"`
	Title            string  `gorm:"not null;index;unique"`
	Description      string  `gorm:"not null"`
	SmallDescription string  `gorm:"not null"`
	Duration         int32   `gorm:"not null"`
	ReleaseDate      int     `gorm:"not null"`
	Country          string  `gorm:"not null"`
	Poster           string  `gorm:"not null"`
	RatingKp         float64 `gorm:"not null"`
}

type FilmGenre struct {
	FilmID  uint
	GenreID uint
	Film    Film
	Genre   Genre
}

type Genre struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique;not null"`
}

func (f *Film) NameAlreadyExist(rep repository.Repository, name string) error {
	if err := rep.Where("Title = ?", name).First(f).Error; err != nil {
		return err
	}
	return nil
}

func (g *Genre) NameAlreadyExist(rep repository.Repository, name string) error {
	if err := rep.Where("name = ?", name).First(g).Error; err != nil {
		return err
	}
	return nil
}

func (f *Film) FindByName(rep repository.Repository, nameFilm string) (*Film, error) {
	if err := rep.Where("Title = ?", nameFilm).First(f).Error; err != nil {
		return nil, appErrors.ErrFilmNotFound
	}
	return f, nil
}

func (f *Film) Create(rep repository.Repository) error {
	if err := rep.Create(f).Error; err != nil {
		return err // бывший ErrInvalidServer
	}
	return nil
}

func (g *Genre) Create(rep repository.Repository) error {
	if err := rep.Create(g).Error; err != nil {
		return err
	}
	return nil
}

func (f *Film) Delete(rep repository.Repository) error {
	if err := rep.Delete(f).Error; err != nil {
		return appErrors.ErrInvalidPassword
	}
	return nil
}
