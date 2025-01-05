package metadata

import (
	"errors"
	"time"
)

type Game struct {
	Name        string
	Summary     string
	ReleaseDate time.Time
	Rating      float64
	Genres      []string
	Cover       string
	Artworks    []string
	Screenshots []string

	Provider   string
	ProviderID string
}

var ErrNotFound = errors.New("game not found")

type Provider interface {
	Find(name string, year int) (*Game, error)
}
