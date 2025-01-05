package steam

import (
	"boyl/server/scan/metadata"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	SearchURL  = "https://store.steampowered.com/api/storesearch"
	DetailsURL = "https://store.steampowered.com/api/appdetails"
	ReviewURL  = "https://store.steampowered.com/appreviews"
	ProviderID = "steam"
)

// why does the fucking date format change ????
func parseShittySteamDate(date string) (time.Time, error) {
	t, err := time.Parse("Jan 2, 2006", date)
	if err != nil {
		t, err = time.Parse("2 Jan, 2006", date)
	}
	return t, err
}

type searchItem struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}
type searchResult struct {
	Items []searchItem `json:"items"`
}

type detailsItem struct {
	Data struct {
		Name             string `json:"name"`
		ShortDescription string `json:"short_description"`
		ReleaseDate      struct {
			Date string `json:"date"`
		} `json:"release_date"`
		Genres []struct {
			Description string `json:"description"`
		} `json:"genres"`
		HeaderImage   string `json:"header_image"`
		BackgroundRaw string `json:"background_raw"`
		Screenshots   []struct {
			PathFull string `json:"path_full"`
		} `json:"screenshots"`
	} `json:"data"`
}
type detailsResult map[string]detailsItem

type reviewResult struct {
	QuerySummary struct {
		TotalPositive int `json:"total_positive"`
		TotalNegative int `json:"total_negative"`
	} `json:"query_summary"`
}

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) search(name string) (*searchResult, error) {
	values := url.Values{}
	values.Set("term", name)
	values.Set("l", "english")
	values.Set("cc", "US")
	fmt.Println(SearchURL + "?" + values.Encode())

	res, err := http.Get(SearchURL + "?" + values.Encode())
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("search failed")
	}

	var result searchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (p *Provider) getDetails(id int) (*detailsItem, error) {
	values := url.Values{}
	values.Set("appids", strconv.Itoa(id))
	fmt.Println(DetailsURL + "?" + values.Encode())

	res, err := http.Get(DetailsURL + "?" + values.Encode())
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("details failed")
	}

	data, _ := io.ReadAll(res.Body)
	fmt.Println(string(data))
	res.Body = io.NopCloser(bytes.NewReader(data))

	var result detailsResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	item, ok := result[strconv.Itoa(id)]
	if !ok {
		return nil, errors.New("no details found")
	}

	return &item, nil
}

func (p *Provider) getReviews(id int) (*reviewResult, error) {
	values := url.Values{}
	values.Set("json", "1")

	res, err := http.Get(ReviewURL + "/" + strconv.Itoa(id) + "?" + values.Encode())
	if err != nil {
		return nil, err
	}

	var result reviewResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func cleanImageURL(u string) string {
	url, err := url.Parse(u)
	if err != nil {
		return u
	}
	url.RawQuery = ""
	return url.String()
}

func (p *Provider) Find(name string, year int) (*metadata.Game, error) {
	s, err := p.search(name)
	if err != nil {
		return nil, err
	}

	var searchItem *searchItem
	for _, item := range s.Items {
		if item.Type == "app" {
			searchItem = &item
			break
		}
	}
	if searchItem == nil {
		return nil, metadata.ErrNotFound
	}

	details, err := p.getDetails(searchItem.ID)
	if err != nil {
		return nil, err
	}
	d := details.Data

	r, err := p.getReviews(searchItem.ID)
	if err != nil {
		return nil, err
	}

	var game metadata.Game

	game.Provider = ProviderID
	game.ProviderID = strconv.Itoa(searchItem.ID)

	game.Name = d.Name
	game.Summary = d.ShortDescription
	game.Cover = cleanImageURL(d.HeaderImage)
	game.Artworks = []string{cleanImageURL(d.BackgroundRaw)}
	game.Rating = float64(r.QuerySummary.TotalPositive) / float64(r.QuerySummary.TotalPositive+r.QuerySummary.TotalNegative) * 100

	game.ReleaseDate, err = parseShittySteamDate(d.ReleaseDate.Date)
	if err != nil {
		return nil, err
	}

	game.Genres = make([]string, len(d.Genres))
	for i, genre := range d.Genres {
		game.Genres[i] = genre.Description
	}

	game.Screenshots = make([]string, len(d.Screenshots))
	for i, screenshot := range d.Screenshots {
		game.Screenshots[i] = cleanImageURL(screenshot.PathFull)
	}

	return &game, nil
}
