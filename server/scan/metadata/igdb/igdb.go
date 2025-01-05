package igdb

import (
	"boyl/server/scan/metadata"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	BaseURL    = "https://api.igdb.com/v4/games"
	ProviderID = "igdb"
)

type Provider struct {
	client *http.Client
}

func NewProvider(clientID, clientSecret string) *Provider {
	return &Provider{client: NewClientCredentialsClient(context.Background(), clientID, clientSecret)}
}

func hqImage(url string) string {
	new := url
	if strings.HasPrefix(url, "//") {
		new = "https:" + url
	}
	return strings.Replace(new, "t_thumb", "t_1080p", 1)
}

type searchResult struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Summary          string  `json:"summary"`
	FirstReleaseDate int     `json:"first_release_date"`
	TotalRating      float64 `json:"total_rating"`
	Genres           []struct {
		Name string `json:"name"`
	} `json:"genres"`
	Cover struct {
		URL string `json:"url"`
	} `json:"cover"`
	Artworks []struct {
		URL string `json:"url"`
	} `json:"artworks"`
	Screenshots []struct {
		URL string `json:"url"`
	} `json:"screenshots"`
}

func (p *Provider) Find(name string, year int) (*metadata.Game, error) {
	var game metadata.Game

	fields := []string{
		"name",
		"summary",
		"first_release_date",
		"total_rating",
		"genres.name",
		"cover.url",
		"artworks.url",
		"screenshots.url",
	}

	yearTime := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	query := fmt.Sprintf(
		`fields %s; search "%s"; where first_release_date > %d & first_release_date < %d; limit 1;`,
		strings.Join(fields, ","),
		name,
		yearTime.Unix(),
		yearTime.AddDate(1, 0, 0).Unix(),
	)
	req, err := http.NewRequest("POST", "https://api.igdb.com/v4/games", strings.NewReader(query))
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var searchResults []searchResult
	if err := json.NewDecoder(resp.Body).Decode(&searchResults); err != nil {
		return nil, err
	}
	if len(searchResults) == 0 {
		return nil, metadata.ErrNotFound
	}

	result := searchResults[0]

	game.Provider = ProviderID
	game.ProviderID = fmt.Sprintf("%d", result.ID)

	game.Name = result.Name
	game.Summary = result.Summary
	game.ReleaseDate = time.Unix(int64(result.FirstReleaseDate), 0)
	game.Rating = result.TotalRating
	game.Cover = hqImage(result.Cover.URL)

	game.Genres = make([]string, len(result.Genres))
	for i, genre := range result.Genres {
		game.Genres[i] = genre.Name
	}
	game.Artworks = make([]string, len(result.Artworks))
	for i, artwork := range result.Artworks {
		game.Artworks[i] = hqImage(artwork.URL)
	}
	game.Screenshots = make([]string, len(result.Screenshots))
	for i, screenshot := range result.Screenshots {
		game.Screenshots[i] = hqImage(screenshot.URL)
	}

	return &game, nil
}
