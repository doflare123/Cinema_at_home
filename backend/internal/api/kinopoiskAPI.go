package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type filmResult struct {
	Title            string
	Description      string
	ShortDescription string
	Poster           string
	RatingKp         float64
	Year             int
	Duration         int
	Genres           []string
	Countries        []string
}

type apiResponse struct {
	Docs []struct {
		Name             string `json:"name"`
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

func SearchFilm(name string, apikey string) (*filmResult, error) {
	endpoint := "https://api.poiskkino.dev/v1.4/movie/search"

	params := url.Values{}
	params.Set("page", "1")
	params.Set("limit", "1")
	params.Set("query", name)

	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("X-API-KEY", apikey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var apiRes apiResponse
	if err := json.NewDecoder(res.Body).Decode(&apiRes); err != nil {
		return nil, err
	}

	if len(apiRes.Docs) == 0 {
		return nil, fmt.Errorf("film not found")
	}

	doc := apiRes.Docs[0]

	result := &filmResult{
		Title:            doc.Name,
		Description:      doc.Description,
		ShortDescription: doc.ShortDescription,
		Poster:           doc.Poster.URL,
		RatingKp:         doc.Rating.Kp,
		Year:             doc.Year,
		Duration:         doc.MovieLength,
	}

	for _, g := range doc.Genres {
		result.Genres = append(result.Genres, g.Name)
	}

	for _, c := range doc.Countries {
		result.Countries = append(result.Countries, c.Name)
	}

	return result, nil
}
