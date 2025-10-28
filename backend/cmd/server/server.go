package main

import (
	"cinema/internal/container"
	"cinema/internal/database"
	"cinema/internal/models"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	cont   container.Container
}

func initServer(cont container.Container) (*Server, error) {
	db := cont.GetRepository()
	logger := cont.GetLogger()
	if cont.GetConfig().AppEnv == "dev" {
		gin.SetMode(gin.DebugMode)
		if err := database.AutoMigDB(db, &models.User{}, &models.Role{}); err != nil {
			logger.Error("Error with auto migration: %s", err)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
		if err := database.MigrationDB(db, logger); err != nil {
			logger.Error("Error with migrations: %s", err)
		}
	}
	err := database.Seeder(db)
	if err != nil {
		logger.Error("Error with seeder", "error", err)
	}

	r := gin.Default()

	s := &Server{
		engine: r,
		cont:   cont,
	}
	s.InitRouters()
	return s, nil
}

func (s *Server) run() error {
	s.cont.GetLogger().Info("Server started")
	return s.engine.Run(s.cont.GetConfig().ServerPort)
}
