package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"errors"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type FranchiseService interface {
	List() ([]dto.FranchiseListItem, error)
	GetByID(id uint) (dto.FranchiseDetailView, error)
	Create(req dto.CreateFranchiseRequest) (dto.FranchiseDetailView, error)
	AddMovie(franchiseID uint, req dto.AddFranchiseMovieRequest) (dto.FranchiseDetailView, error)
}

type franchiseService struct {
	rep repository.Repository
}

func NewFranchiseService(rep repository.Repository) FranchiseService {
	return &franchiseService{rep: rep}
}

func (s *franchiseService) List() ([]dto.FranchiseListItem, error) {
	var franchises []models.Franchise
	if err := s.rep.Preload("Movies").Order("title ASC").Find(&franchises).Error; err != nil {
		return nil, err
	}

	items := make([]dto.FranchiseListItem, 0, len(franchises))
	for _, franchise := range franchises {
		items = append(items, dto.FranchiseListItem{
			ID:          franchise.ID,
			Title:       franchise.Title,
			Description: franchise.Description,
			MoviesCount: len(franchise.Movies),
		})
	}
	return items, nil
}

func (s *franchiseService) GetByID(id uint) (dto.FranchiseDetailView, error) {
	return s.loadDetail(s.rep, id)
}

func (s *franchiseService) Create(req dto.CreateFranchiseRequest) (dto.FranchiseDetailView, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return dto.FranchiseDetailView{}, appErrors.ErrInvalidFranchiseTitle
	}

	var created models.Franchise
	err := s.rep.Transaction(func(tx repository.Repository) error {
		var existing models.Franchise
		if err := tx.Where("title = ?", title).First(&existing).Error; err == nil {
			return appErrors.ErrFranchiseTitleAlreadyExist
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		created = models.Franchise{
			Title:       title,
			Description: strings.TrimSpace(req.Description),
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return dto.FranchiseDetailView{}, err
	}

	return s.loadDetail(s.rep, created.ID)
}

func (s *franchiseService) AddMovie(franchiseID uint, req dto.AddFranchiseMovieRequest) (dto.FranchiseDetailView, error) {
	if req.MovieID == 0 || req.PartNumber <= 0 || req.ReleaseOrder <= 0 || req.ChronologyOrder <= 0 {
		return dto.FranchiseDetailView{}, appErrors.ErrInvalidFranchiseMovieLink
	}

	err := s.rep.Transaction(func(tx repository.Repository) error {
		if _, err := s.loadFranchise(tx, franchiseID); err != nil {
			return err
		}

		var film models.Film
		if err := tx.First(&film, req.MovieID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrFilmNotFound
			}
			return err
		}

		var existing models.FranchiseMovie
		if err := tx.Where("franchise_id = ? AND film_id = ?", franchiseID, req.MovieID).First(&existing).Error; err == nil {
			return appErrors.ErrMovieAlreadyInFranchise
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		link := models.FranchiseMovie{
			FranchiseID:     franchiseID,
			MovieID:         req.MovieID,
			PartNumber:      req.PartNumber,
			ReleaseOrder:    req.ReleaseOrder,
			ChronologyOrder: req.ChronologyOrder,
		}
		return tx.Create(&link).Error
	})
	if err != nil {
		return dto.FranchiseDetailView{}, err
	}

	return s.loadDetail(s.rep, franchiseID)
}

func (s *franchiseService) loadFranchise(rep repository.Repository, id uint) (*models.Franchise, error) {
	var franchise models.Franchise
	if err := rep.First(&franchise, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrFranchiseNotFound
		}
		return nil, err
	}
	return &franchise, nil
}

func (s *franchiseService) loadDetail(rep repository.Repository, id uint) (dto.FranchiseDetailView, error) {
	var franchise models.Franchise
	if err := rep.Preload("Movies.Movie").First(&franchise, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.FranchiseDetailView{}, appErrors.ErrFranchiseNotFound
		}
		return dto.FranchiseDetailView{}, err
	}

	movies := make([]dto.FranchiseMovieView, 0, len(franchise.Movies))
	for _, link := range franchise.Movies {
		movies = append(movies, dto.FranchiseMovieView{
			MovieID:         link.MovieID,
			Title:           link.Movie.Title,
			Poster:          link.Movie.Poster,
			ReleaseDate:     link.Movie.ReleaseDate,
			PartNumber:      link.PartNumber,
			ReleaseOrder:    link.ReleaseOrder,
			ChronologyOrder: link.ChronologyOrder,
		})
	}

	sort.Slice(movies, func(i, j int) bool {
		if movies[i].ChronologyOrder != movies[j].ChronologyOrder {
			return movies[i].ChronologyOrder < movies[j].ChronologyOrder
		}
		if movies[i].ReleaseOrder != movies[j].ReleaseOrder {
			return movies[i].ReleaseOrder < movies[j].ReleaseOrder
		}
		if movies[i].PartNumber != movies[j].PartNumber {
			return movies[i].PartNumber < movies[j].PartNumber
		}
		return movies[i].MovieID < movies[j].MovieID
	})

	return dto.FranchiseDetailView{
		ID:          franchise.ID,
		Title:       franchise.Title,
		Description: franchise.Description,
		Movies:      movies,
	}, nil
}
