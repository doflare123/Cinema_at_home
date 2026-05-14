package main

import (
	"cinema/internal/api"
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/routers"
	"cinema/internal/app/services"
)

func (s *Server) InitRouters() {
	authHandler := handlers.NewAuthHandler(s.cont)
	routers.RegisterAuthRoutes(s.engine, authHandler, s.cont.GetRepository())

	userHandler := handlers.NewUserHandler(s.cont)
	routers.RegisterUserRoutes(s.engine, userHandler, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())

	filmSrv := services.NewFilmService(s.cont.GetLogger(), s.cont.GetRepository(), s.cont.GetConfig().ApiKey)
	filmH := handlers.NewFilmHandler(filmSrv)
	routers.RegisterFilmRoutes(s.engine, filmH, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())

	franchiseSrv := services.NewFranchiseService(s.cont.GetRepository())
	franchiseH := handlers.NewFranchiseHandler(franchiseSrv)
	routers.RegisterFranchiseRoutes(s.engine, franchiseH, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())

	expectationSrv := services.NewExpectationService(s.cont.GetRepository())
	expectationH := handlers.NewExpectationHandler(expectationSrv)
	routers.RegisterExpectationRoutes(s.engine, expectationH, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())

	reviewSrv := services.NewReviewService(s.cont.GetRepository())
	reviewH := handlers.NewReviewHandler(reviewSrv)
	routers.RegisterReviewRoutes(s.engine, reviewH, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())

	proposalSrv := services.NewMovieProposalService(s.cont.GetRepository())
	proposalH := handlers.NewMovieProposalHandler(proposalSrv)
	routers.RegisterMovieProposalRoutes(s.engine, proposalH, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())

	kinopoiskSrv := services.NewKinopoiskService(api.NewKinopoiskClient(s.cont.GetConfig().ApiKey))
	kinopoiskH := handlers.NewKinopoiskHandler(kinopoiskSrv)
	routers.RegisterKinopoiskRoutes(s.engine, kinopoiskH, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())

	weeklyPackSrv := services.NewWeeklyPackService(s.cont.GetRepository())
	weeklyPackH := handlers.NewWeeklyPackHandler(weeklyPackSrv)
	routers.RegisterWeeklyPackRoutes(s.engine, weeklyPackH, s.cont.GetConfig().JWTSecretKey, s.cont.GetRepository())
}
