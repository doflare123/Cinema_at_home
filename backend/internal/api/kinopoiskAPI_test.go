package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKinopoiskClientSearchFilms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/movie/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("query") != "matrix" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("limit") != "2" {
			t.Fatalf("unexpected limit: %s", r.URL.RawQuery)
		}
		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Fatalf("expected api key header")
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"docs": []map[string]interface{}{
				{
					"id":               301,
					"name":             "The Matrix",
					"description":      "Neo story",
					"shortDescription": "Neo",
					"movieLength":      136,
					"year":             1999,
					"poster":           map[string]interface{}{"url": "poster.jpg"},
					"rating":           map[string]interface{}{"kp": 8.5},
					"genres":           []map[string]interface{}{{"name": "sci-fi"}},
					"countries":        []map[string]interface{}{{"name": "US"}},
				},
			},
		})
	}))
	defer server.Close()

	client := &KinopoiskClient{
		APIKey:     "test-key",
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	results, err := client.SearchFilms("matrix", 2)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
	if results[0].SourceID != "301" || results[0].Title != "The Matrix" || results[0].Duration != 136 {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestKinopoiskClientSearchFilmsReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := &KinopoiskClient{BaseURL: server.URL, HTTPClient: server.Client()}
	if _, err := client.SearchFilms("matrix", 1); err == nil {
		t.Fatal("expected status error")
	}
}

func TestSearchFilmWrapperReturnsFirstResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"docs": []map[string]interface{}{
				{
					"id":               10,
					"name":             "First",
					"description":      "First description",
					"shortDescription": "First short",
					"movieLength":      90,
					"year":             2001,
				},
				{
					"id":          11,
					"name":        "Second",
					"movieLength": 100,
					"year":        2002,
				},
			},
		})
	}))
	defer server.Close()

	previousBaseURL := defaultKinopoiskBaseURL
	defaultKinopoiskBaseURL = server.URL
	defer func() { defaultKinopoiskBaseURL = previousBaseURL }()

	result, err := SearchFilm("first", "test-key")
	if err != nil {
		t.Fatalf("search film failed: %v", err)
	}
	if result.SourceID != "10" || result.Title != "First" || result.Duration != 90 {
		t.Fatalf("unexpected wrapper result: %+v", result)
	}
}

func TestSearchFilmWrapperReturnsNotFoundForEmptyDocs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"docs": []map[string]interface{}{}})
	}))
	defer server.Close()

	previousBaseURL := defaultKinopoiskBaseURL
	defaultKinopoiskBaseURL = server.URL
	defer func() { defaultKinopoiskBaseURL = previousBaseURL }()

	if _, err := SearchFilm("missing", "test-key"); err == nil {
		t.Fatal("expected not found error")
	}
}
