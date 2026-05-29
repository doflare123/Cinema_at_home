package services

import (
	"cinema/internal/models"
	"cinema/internal/models/dto"
	"cinema/internal/repository"
	"time"
)

type StatisticsService interface {
	Summary() (dto.StatisticsSummaryView, error)
}

type statisticsService struct {
	rep repository.Repository
}

func NewStatisticsService(rep repository.Repository) StatisticsService {
	return &statisticsService{rep: rep}
}

func (s *statisticsService) Summary() (dto.StatisticsSummaryView, error) {
	summary := dto.StatisticsSummaryView{GeneratedAt: time.Now()}

	if err := s.rep.Model(&models.Film{}).Count(&summary.MoviesTotal).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.Franchise{}).Count(&summary.FranchisesTotal).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.User{}).Where("status = ?", "active").Count(&summary.UsersActive).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}

	if err := s.rep.Model(&models.MovieProposal{}).Where("status = ?", "pending").Count(&summary.ProposalsPending).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.MovieProposal{}).Where("status = ?", "approved").Count(&summary.ProposalsApproved).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.MovieProposal{}).Where("status = ?", "rejected").Count(&summary.ProposalsRejected).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}

	if err := s.rep.Model(&models.ExpectationVote{}).Count(&summary.ExpectationsTotal).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.ExpectationVote{}).Where("vote_type = ?", "score").Count(&summary.ExpectationsNumeric).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.ExpectationVote{}).Where("vote_type = ?", "refuse").Count(&summary.ExpectationsRefuse).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}

	type avgRow struct {
		Avg float64
	}
	var expectationAvg avgRow
	if err := s.rep.Model(&models.ExpectationVote{}).
		Select("COALESCE(AVG(CASE WHEN vote_type = ? THEN score END), 0) AS avg", "score").
		Scan(&expectationAvg).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	summary.ExpectationsAverage = expectationAvg.Avg

	if err := s.rep.Model(&models.Review{}).Count(&summary.ReviewsTotal).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	var reviewAvg avgRow
	if err := s.rep.Model(&models.Review{}).
		Select("COALESCE(AVG(final_score), 0) AS avg").
		Scan(&reviewAvg).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	summary.ReviewsAverage = reviewAvg.Avg

	if err := s.rep.Model(&models.WeeklyPack{}).Count(&summary.WeeklyPacksTotal).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.WeeklyPack{}).Where("status = ?", "voting").Count(&summary.WeeklyPacksVoting).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.WeeklyPack{}).Where("status = ?", "closed").Count(&summary.WeeklyPacksClosed).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.WeeklyPackVote{}).Count(&summary.WeeklyPackVotesTotal).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}
	if err := s.rep.Model(&models.WeeklyPackMovie{}).Count(&summary.WeeklyPackMoviesTotal).Error; err != nil {
		return dto.StatisticsSummaryView{}, err
	}

	return summary, nil
}
