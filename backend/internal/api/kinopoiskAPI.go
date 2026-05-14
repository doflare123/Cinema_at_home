package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var defaultKinopoiskBaseURL = "https://api.poiskkino.dev/v1.4"

type FilmResult struct {
	SourceID         string   `json:"source_id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	ShortDescription string   `json:"small_description"`
	Poster           string   `json:"poster"`
	RatingKp         float64  `json:"rating_kp"`
	Year             int      `json:"release_date"`
	Duration         int      `json:"duration"`
	Genres           []string `json:"genres"`
	Countries        []string `json:"countries"`
}

type KinopoiskClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

type apiResponse struct {
	Docs []struct {
		ID               int    `json:"id"`
		Name             string `json:"name"`
		AlternativeName  string `json:"alternativeName"`
		Description      string `json:"description"`
		ShortDescription string `json:"shortDescription"`
		MovieLength      int    `json:"movieLength"`
		Year             int    `json:"year"`
		Genres           []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Countries []struct {
			Name string `json:"name"`
		} `json:"countries"`
		Poster struct {
			URL string `json:"url"`
		} `json:"poster"`
		Rating struct {
			Kp float64 `json:"kp"`
		} `json:"rating"`
	} `json:"docs"`
}

func NewKinopoiskClient(apiKey string) *KinopoiskClient {
	return &KinopoiskClient{
		APIKey:  apiKey,
		BaseURL: defaultKinopoiskBaseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func SearchFilm(name string, apiKey string) (*FilmResult, error) {
	results, err := NewKinopoiskClient(apiKey).SearchFilms(name, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("film not found")
	}
	return &results[0], nil
}

func (c *KinopoiskClient) SearchFilms(query string, limit int) ([]FilmResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	baseURL := strings.TrimRight(c.BaseURL, "/")
	endpoint := baseURL + "/movie/search"

	params := url.Values{}
	params.Set("page", "1")
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("query", query)

	req, err := http.NewRequest(http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	if strings.TrimSpace(c.APIKey) != "" {
		req.Header.Set("X-API-KEY", c.APIKey)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("kinopoisk api returned status %d", res.StatusCode)
	}

	var apiRes apiResponse
	if err := json.NewDecoder(res.Body).Decode(&apiRes); err != nil {
		return nil, err
	}

	results := make([]FilmResult, 0, len(apiRes.Docs))
	for _, doc := range apiRes.Docs {
		title := strings.TrimSpace(doc.Name)
		if title == "" {
			title = strings.TrimSpace(doc.AlternativeName)
		}
		if title == "" {
			continue
		}

		result := FilmResult{
			SourceID:         fmt.Sprintf("%d", doc.ID),
			Title:            title,
			Description:      doc.Description,
			ShortDescription: doc.ShortDescription,
			Poster:           doc.Poster.URL,
			RatingKp:         doc.Rating.Kp,
			Year:             doc.Year,
			Duration:         doc.MovieLength,
		}

		for _, g := range doc.Genres {
			if name := strings.TrimSpace(g.Name); name != "" {
				result.Genres = append(result.Genres, name)
			}
		}

		for _, c := range doc.Countries {
			if name := strings.TrimSpace(c.Name); name != "" {
				result.Countries = append(result.Countries, name)
			}
		}

		results = append(results, result)
	}

	return results, nil
}
