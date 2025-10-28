package database

import (
	"cinema/config"
	"cinema/internal/models"
	"encoding/json"
	"sync"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Seeder(db *gorm.DB) []error {
	var errs []error

	seeders := []func(*gorm.DB) error{
		roleSeed,
		genreSeed,
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

func genreSeed(db *gorm.DB) error {
	data, err := config.GetJSONFile("genres.json")
	if err != nil {
		return err
	}

	var genres []string
	if err := json.Unmarshal(data, &genres); err != nil {
		return err
	}

	// Загружаем все существующие жанры
	var existing []models.Genre
	if err := db.Find(&existing).Error; err != nil {
		return err
	}

	existingMap := make(map[string]bool)
	for _, g := range existing {
		existingMap[g.Name] = true
	}

	jsonMap := make(map[string]bool)
	for _, name := range genres {
		jsonMap[name] = true
		// Добавляем новые жанры
		if !existingMap[name] {
			if err := db.Create(&models.Genre{Name: name}).Error; err != nil {
				return err
			}
		}
	}

	// Удаляем жанры, которых нет в JSON
	for _, g := range existing {
		if !jsonMap[g.Name] {
			if err := db.Delete(&g).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
