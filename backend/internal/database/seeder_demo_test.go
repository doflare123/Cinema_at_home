package database

import "testing"

func TestDemoCatalogFixtures(t *testing.T) {
	films := demoFilms()
	if len(films) < 4 {
		t.Fatalf("expected at least 4 demo films, got %d", len(films))
	}

	franchises := demoFranchises()
	if len(franchises) < 2 {
		t.Fatalf("expected at least 2 demo franchises, got %d", len(franchises))
	}

	links := demoFranchiseLinks(1, 2, 11, 12, 21, 22)
	if len(links) != 4 {
		t.Fatalf("expected 4 franchise links, got %d", len(links))
	}
}

