package container

import (
	"cinema/config"
	"cinema/internal/database"
	"cinema/internal/logger"

	"gorm.io/gorm"
)

type Container struct {
	AppPort string
	Config  *config.Config
	DB      *gorm.DB
	Logger  logger.Logger
}

func NewContainer() (*Container, error) {
	cfg := config.NewConfig()
	log, err := logger.NewZapLogger()
	if err != nil {
		return nil, err
	}
	db, err := database.InitDB(cfg.DBDsn, log)
	if err != nil {
		log.Error("Problems with DB initialization", "error", err)
		return nil, err
	}
	log.Info("Config and DB initialized", "port", cfg.ServerPort)
	return &Container{
		AppPort: cfg.ServerPort,
		Config:  cfg,
		DB:      db,
		Logger:  log,
	}, nil
}
