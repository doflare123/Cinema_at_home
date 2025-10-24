package main

import (
	"cinema/internal/container"
	"cinema/internal/database"
	"cinema/internal/models"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	cont   *container.Container
}

func initServer(cont *container.Container) (*Server, error) {
	if cont.Config.AppEnv == "dev" {
		gin.SetMode(gin.DebugMode)
		if err := database.AutoMigDB(cont.DB, &models.User{}, &models.Role{}); err != nil {
			cont.Logger.Error("Error with auto migration: %s", err)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
		if err := database.MigrationDB(cont.DB, cont.Logger); err != nil {
			cont.Logger.Error("Error with migrations: %s", err)
		}
	}
	database.Seeder(cont.DB)

	r := gin.Default()

	s := &Server{
		engine: r,
		cont:   cont,
	}
	s.InitRouters()
	return s, nil
}

func (s *Server) run() error {
	s.cont.Logger.Info("Server started")
	return s.engine.Run(s.cont.Config.ServerPort)
}
