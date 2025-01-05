package gog

import (
	"boyl/server/scan/metadata"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

const (
	BaseURL    = "https://catalog.gog.com/v1/catalog"
	GameURL    = "https://www.gog.com/en/game"
	ProviderID = "gog"
)

func getDescription(slug string) (string, error) {
	res, err := http.Get(GameURL + "/" + slug)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", err
	}

	description := doc.Find("div.description")
	description.Find("*").Each(func(i int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		if class == "description__copyrights" || class == "module" {
			s.Remove()
		}

		s.RemoveAttr("class")
	})

	childrenHTML, err := description.Html()
	if err != nil {
		return "", err
	}

	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	childrenHTML, err = m.String("text/html", childrenHTML)
	if err != nil {
		return "", err
	}

	return childrenHTML, nil
}

var dateFormat = "2006.01.02"

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func hqImage(url string) string {
	return strings.ReplaceAll(url, "_{formatter}", "")
}

type searchResult struct {
	Products []searchProduct `json:"products"`
}

type searchProduct struct {
	ID            string `json:"id"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	ReleaseDate   string `json:"releaseDate"`
	ReviewsRating int    `json:"reviewsRating"`
	Genres        []struct {
		Name string `json:"name"`
	}
	CoverHorizontal string   `json:"coverHorizontal"`
	CoverVertical   string   `json:"coverVertical"`
	Screenshots     []string `json:"screenshots"`
}

func (p *Provider) Find(name string, year int) (*metadata.Game, error) {
	form := url.Values{}
	form.Set("limit", "1")
	form.Set("productType", "in:game,pack")
	form.Set("releaseDate", fmt.Sprintf("beetween:%d,%d", year-1, year+1))
	form.Set("query", "like:"+name)

	res, err := http.Get(BaseURL + "?" + form.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var result searchResult
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Products) == 0 {
		return nil, metadata.ErrNotFound
	}
	product := result.Products[0]

	var game metadata.Game

	game.Provider = ProviderID
	game.ProviderID = product.ID

	game.Name = product.Title
	game.Rating = float64(product.ReviewsRating) * 2
	game.Cover = hqImage(product.CoverVertical)
	game.Artworks = []string{hqImage(product.CoverHorizontal)}

	description, err := getDescription(product.Slug)
	if err != nil {
		return nil, err
	}
	game.Summary = description

	game.ReleaseDate, err = time.Parse(dateFormat, product.ReleaseDate)
	if err != nil {
		return nil, err
	}

	game.Genres = make([]string, len(product.Genres))
	for i, genre := range product.Genres {
		game.Genres[i] = genre.Name
	}

	game.Screenshots = make([]string, len(product.Screenshots))
	for i, screenshot := range product.Screenshots {
		game.Screenshots[i] = hqImage(screenshot)
	}

	return &game, nil
}
