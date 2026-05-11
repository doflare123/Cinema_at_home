package main

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/routers"
	"cinema/internal/app/services"
)

func (s *Server) InitRouters() {
	authHandler := handlers.NewAuthHandler(s.cont)
	routers.RegisterAuthRoutes(s.engine, authHandler)

	userHandler := handlers.NewUserHandler(s.cont)
	routers.RegisterUserRoutes(s.engine, userHandler, s.cont.GetConfig().JWTSecretKey)

	filmSrv := services.NewFilmService(s.cont.GetLogger(), s.cont.GetRepository(), s.cont.GetConfig().ApiKey)
	filmH := handlers.NewFilmHandler(filmSrv)
	routers.RegisterFilmRoutes(s.engine, filmH, s.cont.GetConfig().JWTSecretKey)
}
