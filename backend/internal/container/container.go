package container

import (
	"cinema/config"
	"cinema/internal/database"
	"cinema/internal/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Container struct {
	Config *config.Config
	DB     *gorm.DB
	Logger *zap.Logger
}

func NewContainer() (*Container, error) {
	cfg := config.NewConfig()
	log, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}
	db, err := database.InitDB(cfg.DBDsn, log)
	if err != nil {
		log.Sugar().Error("Problems with DB initialization: %s", err.Error())
		return nil, err
	}
	log.Sugar().Info("Config and DB initialized")
	return &Container{
		Config: cfg,
		DB:     db,
		Logger: log,
	}, nil
}
