package database

import (
	"cinema/internal/logger"
	"cinema/internal/repository"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func AutoMigDB(db repository.Repository, models ...interface{}) error {
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return err
		}
	}
	return nil
}

func MigrationDB(db repository.Repository, logger logger.Logger) error {
	sqlDB, err := db.GetSQLDB()
	if err != nil {
		return err
	}

	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		return err
	}

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

func resolveMigrationsPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(wd, "migrations"),
		filepath.Join(wd, "backend", "migrations"),
	}
	for _, candidate := range candidates {
		if _, statErr := os.Stat(candidate); statErr == nil {
			return "file://" + filepath.ToSlash(candidate), nil
		}
	}
	return "", fmt.Errorf("migrations directory not found from working dir: %s", wd)
}
