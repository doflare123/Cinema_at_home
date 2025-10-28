package main

import (
	"cinema/config"
	"cinema/internal/app/utils"
	"cinema/internal/container"
	"cinema/internal/database"
	"cinema/internal/logger"
	"log"
)

func main() {
	utils.RegisterPasswordValidator()
	conf := config.NewConfig()
	logger, err := logger.NewZapLogger()
	if err != nil {
		log.Fatal(err)
	}
	logger.Info("Logger initialized")

	rep := database.NewFilmRepository(logger, *conf)
	container, err := container.NewContainer(rep, logger, *conf)
	if err != nil {
		panic(err)
	}

	serv, err := initServer(container)
	if err != nil {
		logger.Error("Error with server initialization: %s", err)
	}
	if err := serv.run(); err != nil {
		panic(err)
	}
}
