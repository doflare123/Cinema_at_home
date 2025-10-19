package database

import (
	"cinema/internal/models"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Seeder(db *gorm.DB) []error {
	var errs []error

	seeders := []func(*gorm.DB) error{
		roleSeed,
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(seeders))

	for _, seed := range seeders {
		wg.Add(1)
		go func(seed func(*gorm.DB) error) {
			defer wg.Done()
			if err := seed(db); err != nil {
				errCh <- err
			}
		}(seed)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func roleSeed(db *gorm.DB) error {
	roles := []models.Role{
		{Name: "user"},
		{Name: "admin"},
	}
	for _, r := range roles {
		err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&r).Error
		if err != nil {
			return err
		}
	}
	return nil
}
