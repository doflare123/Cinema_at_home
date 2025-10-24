package main

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/routers"
	"cinema/internal/app/services"
)

func (s *Server) InitRouters() {
	// РЕГИСТРАЦИЯ СЕРВИСОВ И ХЭНДЛЕРОВ АВТОРИЗАЦИИ
	registerService := services.NewRegisterServices(s.cont.DB, s.cont.Logger)
	registerHandler := handlers.NewRegisterHandler(registerService)
	authService := services.NewAuthServices(s.cont.DB, s.cont.Logger, s.cont.Config.JWTSecretKey)
	authHandler := handlers.NewAuthHandler(authService)
	routers.RegisterAuthRoutes(s.engine, registerHandler, authHandler)
}
