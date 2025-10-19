package main

import (
	"cinema/internal/container"
	"cinema/internal/database"
	"cinema/internal/models"

	"github.com/gin-gonic/gin"
)

func main() {
	cont, err := container.NewContainer()
	if err != nil {
		panic(err)
	}
	if cont.Config.AppEnv == "dev" {
		gin.SetMode(gin.DebugMode)
		if err := database.AutoMigDB(cont.DB, &models.User{}, &models.Role{}); err != nil {
			cont.Logger.Sugar().Error("Error with auto migration: %s", err)
		}
	} else {
		gin.SetMode(gin.ReleaseMode)
		if err := database.MigrationDB(cont.DB, cont.Logger); err != nil {
			cont.Logger.Sugar().Error("Error with migrations: %s", err)
		}
	}
	database.Seeder(cont.DB)
	r := gin.Default()

	if err := r.Run(); err != nil {
		panic(err)
	} else {
		cont.Logger.Sugar().Info("Server started")
	}
}
