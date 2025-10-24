package database

import (
	"cinema/internal/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string, logger logger.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	logger.Info("connected to DB")
	return db, nil
}
