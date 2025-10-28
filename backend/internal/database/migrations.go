package database

import (
	"cinema/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func AutoMigDB(db Repository, models ...interface{}) error {
	if err := db.AutoMigrate(models); err != nil {
		return err
	}
	return nil
}

func MigrationDB(db Repository, logger logger.Logger) error {
	sqlDB, err := db.GetSQLDB()
	if err != nil {
		return err
	}
	migrationsPath := "file://D:/Приколюхи/Cinema_at_home/backend/migrations"
	driver, err := migratepg.WithInstance(sqlDB, &migratepg.Config{})
	if err != nil {
		return err
	}
	migrat, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}
	if err := migrat.Up(); err != nil {
		if err == migrate.ErrNoChange {
			logger.Info("No new migrations to apply")
		} else {
			return err
		}
	}
	logger.Info("Migrations done")
	return nil
}
