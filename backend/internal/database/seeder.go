package database

import (
	"cinema/config"
	"cinema/internal/models"
	"cinema/internal/repository"
	"encoding/json"
	"sync"

	"gorm.io/gorm/clause"
)

func Seeder(db repository.Repository) []error {
	var errs []error

	seeders := []func(repository.Repository) error{
		roleSeed,
		genreSeed,
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(seeders))

	for _, seed := range seeders {
		wg.Add(1)
		go func(seed func(repository.Repository) error) {
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

func DemoCatalogSeed(db repository.Repository) error {
	films := demoFilms()
	for i := range films {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "title"}},
			DoNothing: true,
		}).Create(&films[i]).Error; err != nil {
			return err
		}
	}

	franchises := demoFranchises()
	for i := range franchises {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "title"}},
			DoNothing: true,
		}).Create(&franchises[i]).Error; err != nil {
			return err
		}
	}

	var lotr models.Franchise
	if err := db.Where("title = ?", "The Lord of the Rings").First(&lotr).Error; err != nil {
		return err
	}
	var dune models.Franchise
	if err := db.Where("title = ?", "Dune").First(&dune).Error; err != nil {
		return err
	}

	var lotr1, lotr2, dune1, dune2 models.Film
	if err := db.Where("title = ?", "The Fellowship of the Ring").First(&lotr1).Error; err != nil {
		return err
	}
	if err := db.Where("title = ?", "The Two Towers").First(&lotr2).Error; err != nil {
		return err
	}
	if err := db.Where("title = ?", "Dune: Part One").First(&dune1).Error; err != nil {
		return err
	}
	if err := db.Where("title = ?", "Dune: Part Two").First(&dune2).Error; err != nil {
		return err
	}

	links := demoFranchiseLinks(lotr.ID, dune.ID, lotr1.ID, lotr2.ID, dune1.ID, dune2.ID)
	for i := range links {
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "franchise_id"}, {Name: "film_id"}},
			DoNothing: true,
		}).Create(&links[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

func demoFilms() []models.Film {
	return []models.Film{
		{
			Title:            "The Fellowship of the Ring",
			Description:      "A young hobbit begins a journey to destroy a dangerous ring.",
			SmallDescription: "Start of the LOTR trilogy.",
			Duration:         178,
			ReleaseDate:      2001,
			Country:          "New Zealand",
			Poster:           "https://example.com/posters/lotr1.jpg",
			RatingKp:         8.8,
		},
		{
			Title:            "The Two Towers",
			Description:      "The fellowship is broken, but the mission continues across Middle-earth.",
			SmallDescription: "Second LOTR film.",
			Duration:         179,
			ReleaseDate:      2002,
			Country:          "New Zealand",
			Poster:           "https://example.com/posters/lotr2.jpg",
			RatingKp:         8.7,
		},
		{
			Title:            "Dune: Part One",
			Description:      "Paul Atreides and his family move to Arrakis, the desert planet.",
			SmallDescription: "Beginning of the modern Dune adaptation.",
			Duration:         155,
			ReleaseDate:      2021,
			Country:          "USA",
			Poster:           "https://example.com/posters/dune1.jpg",
			RatingKp:         8.0,
		},
		{
			Title:            "Dune: Part Two",
			Description:      "Paul embraces his destiny and rises against House Harkonnen.",
			SmallDescription: "Continuation of Dune saga.",
			Duration:         166,
			ReleaseDate:      2024,
			Country:          "USA",
			Poster:           "https://example.com/posters/dune2.jpg",
			RatingKp:         8.4,
		},
	}
}

func demoFranchises() []models.Franchise {
	return []models.Franchise{
		{Title: "The Lord of the Rings", Description: "Epic fantasy trilogy based on J.R.R. Tolkien."},
		{Title: "Dune", Description: "Sci-fi saga on Arrakis and the rise of Paul Atreides."},
	}
}

func demoFranchiseLinks(lotrID, duneID, lotr1ID, lotr2ID, dune1ID, dune2ID uint) []models.FranchiseMovie {
	return []models.FranchiseMovie{
		{FranchiseID: lotrID, MovieID: lotr1ID, PartNumber: 1, ReleaseOrder: 1, ChronologyOrder: 1},
		{FranchiseID: lotrID, MovieID: lotr2ID, PartNumber: 2, ReleaseOrder: 2, ChronologyOrder: 2},
		{FranchiseID: duneID, MovieID: dune1ID, PartNumber: 1, ReleaseOrder: 1, ChronologyOrder: 1},
		{FranchiseID: duneID, MovieID: dune2ID, PartNumber: 2, ReleaseOrder: 2, ChronologyOrder: 2},
	}
}

func roleSeed(db repository.Repository) error {
	roles := []models.Role{
		{Name: "member"},
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

func genreSeed(db repository.Repository) error {
	data, err := config.GetJSONFile("genres.json")
	if err != nil {
		return err
	}

	var genres []string
	if err := json.Unmarshal(data, &genres); err != nil {
		return err
	}

	var existing []models.Genre
	if err := db.Find(&existing).Error; err != nil {
		return err
	}

	existingMap := make(map[string]bool)
	for _, g := range existing {
		existingMap[g.Name] = true
	}

	for _, name := range genres {
		if !existingMap[name] {
			if err := db.Create(&models.Genre{Name: name}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
