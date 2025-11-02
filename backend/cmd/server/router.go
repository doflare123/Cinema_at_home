package main

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/routers"
)

func (s *Server) InitRouters() {
	// РЕГИСТРАЦИЯ СЕРВИСОВ И ХЭНДЛЕРОВ АВТОРИЗАЦИИ
	authHandler := handlers.NewAuthHandler(s.cont)
	routers.RegisterAuthRoutes(s.engine, authHandler)
	// РЕГИСТРАЦИ СЕРВИСОВ И ХЭНДЛЕРОВ ИНФОРМАЦИИ ПОЛЬЗОВАТЕЛЯ
	userHandler := handlers.NewUserHandler(s.cont)
	routers.RegisterUserRoutes(s.engine, userHandler)
}
