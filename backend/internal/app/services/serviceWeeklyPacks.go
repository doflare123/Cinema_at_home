package services

import (
	appErrors "cinema/internal/errors"
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"errors"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	weeklyPackStatusDraft    = "draft"
	weeklyPackStatusVoting   = "voting"
	weeklyPackStatusClosed   = "closed"
	weeklyPackStatusArchived = "archived"
)

var weeklyPackVoteLimits = map[int]int{
	3:  1,
	2:  2,
	1:  2,
	-2: 1,
}

type WeeklyPackService interface {
	List() ([]dto.WeeklyPackListItem, error)
	GetByID(id uint) (dto.WeeklyPackDetailView, error)
	UpsertVote(packID, userID uint, req dto.UpsertWeeklyPackVoteRequest) (dto.WeeklyPackUserVoteItem, error)
	GetUserVotes(packID, userID uint) (dto.WeeklyPackUserVotesView, error)
	Create(createdByUserID uint, req dto.CreateWeeklyPackRequest) (dto.WeeklyPackDetailView, error)
	AddMovie(packID uint, req dto.AddWeeklyPackMovieRequest) (dto.WeeklyPackDetailView, error)
	UpdateStatus(packID uint, req dto.UpdateWeeklyPackStatusRequest) (dto.WeeklyPackDetailView, error)
}

type weeklyPackService struct {
	rep repository.Repository
}

func NewWeeklyPackService(rep repository.Repository) WeeklyPackService {
	return &weeklyPackService{rep: rep}
}

func (s *weeklyPackService) List() ([]dto.WeeklyPackListItem, error) {
	var packs []models.WeeklyPack
	if err := s.rep.Preload("Movies").Preload("Votes").Order("created_at DESC, id DESC").Find(&packs).Error; err != nil {
		return nil, err
	}

	items := make([]dto.WeeklyPackListItem, 0, len(packs))
	for _, pack := range packs {
		items = append(items, dto.WeeklyPackListItem{
			ID:              pack.ID,
			Name:            pack.Name,
			Status:          pack.Status,
			StartsAt:        pack.StartsAt,
			EndsAt:          pack.EndsAt,
			CreatedByUserID: pack.CreatedByUserID,
			MoviesCount:     len(pack.Movies),
			VotesCount:      len(pack.Votes),
		})
	}
	return items, nil
}

func (s *weeklyPackService) GetByID(id uint) (dto.WeeklyPackDetailView, error) {
	return s.loadDetail(s.rep, id)
}

func (s *weeklyPackService) UpsertVote(packID, userID uint, req dto.UpsertWeeklyPackVoteRequest) (dto.WeeklyPackUserVoteItem, error) {
	if req.MovieID == 0 {
		return dto.WeeklyPackUserVoteItem{}, appErrors.ErrWeeklyPackMovieNotFound
	}
	score, err := validateWeeklyPackVoteScore(req.Score)
	if err != nil {
		return dto.WeeklyPackUserVoteItem{}, err
	}

	err = s.rep.Transaction(func(tx repository.Repository) error {
		pack, err := s.loadPackForUpdate(tx, packID)
		if err != nil {
			return err
		}
		if pack.Status != weeklyPackStatusVoting {
			return appErrors.ErrWeeklyPackVotingClosed
		}

		var packMovie models.WeeklyPackMovie
		if err := tx.Where("pack_id = ? AND film_id = ?", packID, req.MovieID).First(&packMovie).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrWeeklyPackMovieNotFound
			}
			return err
		}

		var existing models.WeeklyPackVote
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("pack_id = ? AND film_id = ? AND user_id = ?", packID, req.MovieID, userID).
			First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var userVotes []models.WeeklyPackVote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("pack_id = ? AND user_id = ?", packID, userID).
			Find(&userVotes).Error; err != nil {
			return err
		}

		counts := make(map[int]int, len(weeklyPackVoteLimits))
		for _, vote := range userVotes {
			counts[vote.Score]++
		}
		if existing.ID != 0 {
			counts[existing.Score]--
		}

		if limit, limited := weeklyPackVoteLimits[score]; limited && counts[score] >= limit {
			return appErrors.ErrWeeklyPackVoteLimitExceeded
		}
		slot, err := chooseLimitSlot(score, userVotes, existing.ID)
		if err != nil {
			return err
		}

		if existing.ID == 0 {
			newVote := models.WeeklyPackVote{
				PackID:    packID,
				MovieID:   req.MovieID,
				UserID:    userID,
				Score:     score,
				LimitSlot: slot,
			}
			return tx.Create(&newVote).Error
		}

		existing.Score = score
		existing.LimitSlot = slot
		return tx.Save(&existing).Error
	})
	if err != nil {
		return dto.WeeklyPackUserVoteItem{}, err
	}

	return dto.WeeklyPackUserVoteItem{
		MovieID: req.MovieID,
		Score:   score,
	}, nil
}

func chooseLimitSlot(score int, userVotes []models.WeeklyPackVote, currentVoteID uint) (*int, error) {
	limit, limited := weeklyPackVoteLimits[score]
	if !limited {
		return nil, nil
	}

	used := make(map[int]bool, limit)
	for _, vote := range userVotes {
		if vote.ID == currentVoteID || vote.Score != score || vote.LimitSlot == nil {
			continue
		}
		used[*vote.LimitSlot] = true
	}
	for i := 1; i <= limit; i++ {
		if !used[i] {
			slot := i
			return &slot, nil
		}
	}
	return nil, appErrors.ErrWeeklyPackVoteLimitExceeded
}

func (s *weeklyPackService) GetUserVotes(packID, userID uint) (dto.WeeklyPackUserVotesView, error) {
	if _, err := s.loadPack(s.rep, packID); err != nil {
		return dto.WeeklyPackUserVotesView{}, err
	}

	var votes []models.WeeklyPackVote
	if err := s.rep.Where("pack_id = ? AND user_id = ?", packID, userID).Order("film_id ASC").Find(&votes).Error; err != nil {
		return dto.WeeklyPackUserVotesView{}, err
	}

	items := make([]dto.WeeklyPackUserVoteItem, 0, len(votes))
	for _, vote := range votes {
		items = append(items, dto.WeeklyPackUserVoteItem{
			MovieID: vote.MovieID,
			Score:   vote.Score,
		})
	}

	return dto.WeeklyPackUserVotesView{
		PackID: packID,
		Votes:  items,
	}, nil
}

func (s *weeklyPackService) Create(createdByUserID uint, req dto.CreateWeeklyPackRequest) (dto.WeeklyPackDetailView, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return dto.WeeklyPackDetailView{}, appErrors.ErrInvalidWeeklyPackName
	}
	if req.StartsAt != nil && req.EndsAt != nil && !req.EndsAt.After(*req.StartsAt) {
		return dto.WeeklyPackDetailView{}, appErrors.ErrInvalidWeeklyPackSchedule
	}

	pack := models.WeeklyPack{
		Name:            name,
		Status:          weeklyPackStatusDraft,
		StartsAt:        req.StartsAt,
		EndsAt:          req.EndsAt,
		CreatedByUserID: createdByUserID,
	}
	if err := s.rep.Create(&pack).Error; err != nil {
		return dto.WeeklyPackDetailView{}, err
	}

	return s.loadDetail(s.rep, pack.ID)
}

func (s *weeklyPackService) AddMovie(packID uint, req dto.AddWeeklyPackMovieRequest) (dto.WeeklyPackDetailView, error) {
	if req.MovieID == 0 {
		return dto.WeeklyPackDetailView{}, appErrors.ErrWeeklyPackMovieNotFound
	}

	err := s.rep.Transaction(func(tx repository.Repository) error {
		pack, err := s.loadPackForUpdate(tx, packID)
		if err != nil {
			return err
		}
		if pack.Status != weeklyPackStatusDraft {
			return appErrors.ErrWeeklyPackMoviesLocked
		}

		var film models.Film
		if err := tx.First(&film, req.MovieID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return appErrors.ErrFilmNotFound
			}
			return err
		}

		var existing models.WeeklyPackMovie
		if err := tx.Where("pack_id = ? AND film_id = ?", packID, req.MovieID).First(&existing).Error; err == nil {
			return appErrors.ErrWeeklyPackMovieAlreadyExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		return tx.Create(&models.WeeklyPackMovie{
			PackID:    packID,
			MovieID:   req.MovieID,
			SortOrder: req.SortOrder,
		}).Error
	})
	if err != nil {
		return dto.WeeklyPackDetailView{}, err
	}

	return s.loadDetail(s.rep, packID)
}

func (s *weeklyPackService) UpdateStatus(packID uint, req dto.UpdateWeeklyPackStatusRequest) (dto.WeeklyPackDetailView, error) {
	nextStatus := strings.TrimSpace(req.Status)
	if !isWeeklyPackStatus(nextStatus) {
		return dto.WeeklyPackDetailView{}, appErrors.ErrInvalidWeeklyPackStatus
	}

	err := s.rep.Transaction(func(tx repository.Repository) error {
		pack, err := s.loadPackForUpdate(tx, packID)
		if err != nil {
			return err
		}
		if !canTransitionWeeklyPackStatus(pack.Status, nextStatus) {
			return appErrors.ErrWeeklyPackStatusTransition
		}
		if nextStatus == weeklyPackStatusVoting {
			var count int64
			if err := tx.Model(&models.WeeklyPackMovie{}).Where("pack_id = ?", packID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return appErrors.ErrWeeklyPackMustHaveMovies
			}
		}

		updates := map[string]interface{}{
			"status": nextStatus,
		}
		now := time.Now()
		if nextStatus == weeklyPackStatusVoting && pack.StartsAt == nil {
			updates["starts_at"] = now
		}
		if nextStatus == weeklyPackStatusClosed && pack.EndsAt == nil {
			updates["ends_at"] = now
		}

		return tx.Model(&models.WeeklyPack{}).Where("id = ?", packID).Updates(updates).Error
	})
	if err != nil {
		return dto.WeeklyPackDetailView{}, err
	}

	return s.loadDetail(s.rep, packID)
}

func (s *weeklyPackService) loadDetail(rep repository.Repository, id uint) (dto.WeeklyPackDetailView, error) {
	var pack models.WeeklyPack
	if err := rep.
		Preload("Movies", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, film_id ASC")
		}).
		Preload("Movies.Movie").
		Preload("Votes", func(db *gorm.DB) *gorm.DB {
			return db.Order("vote_value DESC, film_id ASC, user_id ASC")
		}).
		Preload("Votes.User").
		First(&pack, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return dto.WeeklyPackDetailView{}, appErrors.ErrWeeklyPackNotFound
		}
		return dto.WeeklyPackDetailView{}, err
	}

	type aggregate struct {
		total int
		votes []dto.WeeklyPackVoteBreakdownItem
	}

	aggregates := make(map[uint]*aggregate, len(pack.Movies))
	for _, movie := range pack.Movies {
		aggregates[movie.MovieID] = &aggregate{
			votes: make([]dto.WeeklyPackVoteBreakdownItem, 0),
		}
	}

	for _, vote := range pack.Votes {
		agg, ok := aggregates[vote.MovieID]
		if !ok {
			continue
		}
		agg.total += vote.Score
		agg.votes = append(agg.votes, dto.WeeklyPackVoteBreakdownItem{
			UserID:      vote.UserID,
			Username:    vote.User.Username,
			DisplayName: vote.User.DisplayName,
			Score:       vote.Score,
		})
	}

	movies := make([]dto.WeeklyPackMovieView, 0, len(pack.Movies))
	for _, movie := range pack.Movies {
		agg := aggregates[movie.MovieID]
		view := dto.WeeklyPackMovieView{
			MovieID:     movie.MovieID,
			Title:       movie.Movie.Title,
			Poster:      movie.Movie.Poster,
			ReleaseDate: movie.Movie.ReleaseDate,
			SortOrder:   movie.SortOrder,
			Votes:       agg.votes,
			ScoreTotal:  agg.total,
		}

		for _, vote := range agg.votes {
			switch vote.Score {
			case 3:
				view.Plus3Count++
			case 2:
				view.Plus2Count++
			case 1:
				view.Plus1Count++
			case 0:
				view.ZeroCount++
			case -2:
				view.Minus2Count++
			}
		}
		sort.Slice(view.Votes, func(i, j int) bool {
			if view.Votes[i].Score != view.Votes[j].Score {
				return view.Votes[i].Score > view.Votes[j].Score
			}
			if view.Votes[i].DisplayName != view.Votes[j].DisplayName {
				return view.Votes[i].DisplayName < view.Votes[j].DisplayName
			}
			return view.Votes[i].UserID < view.Votes[j].UserID
		})
		movies = append(movies, view)
	}

	sort.Slice(movies, func(i, j int) bool {
		if movies[i].ScoreTotal != movies[j].ScoreTotal {
			return movies[i].ScoreTotal > movies[j].ScoreTotal
		}
		if movies[i].SortOrder != movies[j].SortOrder {
			return movies[i].SortOrder < movies[j].SortOrder
		}
		return movies[i].MovieID < movies[j].MovieID
	})

	return dto.WeeklyPackDetailView{
		ID:              pack.ID,
		Name:            pack.Name,
		Status:          pack.Status,
		StartsAt:        pack.StartsAt,
		EndsAt:          pack.EndsAt,
		CreatedByUserID: pack.CreatedByUserID,
		Movies:          movies,
	}, nil
}

func (s *weeklyPackService) loadPack(rep repository.Repository, id uint) (*models.WeeklyPack, error) {
	var pack models.WeeklyPack
	if err := rep.First(&pack, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrWeeklyPackNotFound
		}
		return nil, err
	}
	return &pack, nil
}

func (s *weeklyPackService) loadPackForUpdate(rep repository.Repository, id uint) (*models.WeeklyPack, error) {
	var pack models.WeeklyPack
	if err := rep.Clauses(clause.Locking{Strength: "UPDATE"}).First(&pack, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrWeeklyPackNotFound
		}
		return nil, err
	}
	return &pack, nil
}

func validateWeeklyPackVoteScore(score *int) (int, error) {
	if score == nil {
		return 0, appErrors.ErrInvalidWeeklyPackVoteScore
	}
	switch *score {
	case 3, 2, 1, 0, -2:
		return *score, nil
	default:
		return 0, appErrors.ErrInvalidWeeklyPackVoteScore
	}
}

func isWeeklyPackStatus(status string) bool {
	switch status {
	case weeklyPackStatusDraft, weeklyPackStatusVoting, weeklyPackStatusClosed, weeklyPackStatusArchived:
		return true
	default:
		return false
	}
}

func canTransitionWeeklyPackStatus(currentStatus, nextStatus string) bool {
	if currentStatus == nextStatus {
		return true
	}
	switch currentStatus {
	case weeklyPackStatusDraft:
		return nextStatus == weeklyPackStatusVoting || nextStatus == weeklyPackStatusArchived
	case weeklyPackStatusVoting:
		return nextStatus == weeklyPackStatusClosed
	case weeklyPackStatusClosed:
		return nextStatus == weeklyPackStatusArchived
	default:
		return false
	}
}
