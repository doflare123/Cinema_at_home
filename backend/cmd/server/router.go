package main

import (
	"cinema/internal/app/http/handlers"
	"cinema/internal/app/http/routers"
)

func (s *Server) InitRouters() {
	// РЕГИСТРАЦИЯ СЕРВИСОВ И ХЭНДЛЕРОВ АВТОРИЗАЦИИ
	userHandler := handlers.NewAuthHandler(s.cont)
	routers.RegisterAuthRoutes(s.engine, userHandler)
}
