package scan

import (
	"boyl/server/scan/metadata"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/filesystem"
)

const ScanStatusID = "status1scanning"

var extensions = []string{
	"zip",
	"tar.gz",
	"tar.zst",
	"7z",
	"rar",
}

func isArchive(path string) bool {
	for _, ext := range extensions {
		if filepath.Ext(path) == "."+ext {
			return true
		}
	}
	return false
}

type Scanner struct {
	path             string
	meta             []metadata.Provider
	scanning         bool
	app              core.App
	gamesCollection  *core.Collection
	statusCollection *core.Collection
}

type Match struct {
	Path             string
	FilenameMetadata *FilenameMetadata
	Game             *metadata.Game
}
type Missing struct {
	Path             string
	FilenameMetadata *FilenameMetadata
}

type Result struct {
	Invalid      []string
	Missing      []Missing
	Matches      []Match
	SkipNotFound []string
}

func NewScanner(path string, meta []metadata.Provider, app core.App, gamesCollection *core.Collection, statusCollection *core.Collection) *Scanner {
	return &Scanner{
		path:             path,
		meta:             meta,
		app:              app,
		gamesCollection:  gamesCollection,
		statusCollection: statusCollection,
	}
}

func (s *Scanner) IsScanning() bool {
	return s.scanning
}

func (s *Scanner) GetTotal(skip []string) (int, error) {
	var total int

	err := filepath.WalkDir(s.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isArchive(path) {
			return nil
		}
		for _, skipPath := range skip {
			if path == skipPath {
				return nil
			}
		}
		total++
		return nil
	})
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Scanner) scan(skip []string, progress chan<- bool) (*Result, error) {
	var result Result
	err := filepath.WalkDir(s.path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isArchive(path) {
			return nil
		}
		for _, skipPath := range skip {
			if path == skipPath {
				return nil
			}
		}

		meta, err := ParseFilename(filepath.Base(path))
		if err != nil {
			result.Invalid = append(result.Invalid, path)
			return nil
		}

		var game *metadata.Game
		for _, provider := range s.meta {
			g, err := provider.Find(meta.Name, meta.Year)
			if err == metadata.ErrNotFound {
				continue
			}
			if err != nil {
				return err
			}
			game = g
			break
		}

		if game == nil {
			result.Missing = append(result.Missing, Missing{
				Path:             path,
				FilenameMetadata: meta,
			})
			return nil
		}

		result.Matches = append(result.Matches, Match{
			Path:             path,
			FilenameMetadata: meta,
			Game:             game,
		})

		progress <- true

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *Scanner) Update(ctx context.Context) error {
	if s.scanning {
		return errors.New("scanning in progress")
	}
	s.scanning = true
	defer func() {
		s.scanning = false
	}()

	status, err := s.app.FindFirstRecordByData("status", "id", ScanStatusID)
	if err != nil {
		status = core.NewRecord(s.statusCollection)
		status.Set("id", ScanStatusID)
	}
	status.Set("name", "Scanning in progress")
	status.Set("text", "Preparing to scan")
	err = s.app.Save(status)
	if err != nil {
		return err
	}

	defer func() {
		s.app.Delete(status)
	}()

	var paths []string
	games, err := s.app.FindAllRecords("games")
	if err != nil {
		return err
	}
	for _, game := range games {
		path := game.GetString("path")

		if _, err := os.Stat(path); os.IsNotExist(err) {
			s.app.Logger().Error("file for game not found", "path", path, "name", game.GetString("name"))
			game.Set("status", "deleted")
			if err := s.app.Save(game); err != nil {
				return err
			}
		}

		status := game.GetString("status")
		if status == "found" {
			paths = append(paths, path)
		}
	}

	total, err := s.GetTotal(paths)
	if err != nil {
		return err
	}
	status.Set("text", "Scanning")
	status.Set("total", total)
	s.app.Save(status)

	progress := make(chan bool)
	go func() {
		var i int
		for range progress {
			i++
			status.Set("current", i)
			s.app.Save(status)
		}
	}()

	result, err := s.scan(paths, progress)
	close(progress)
	if err != nil {
		return err
	}

	status.Set("text", "Applying changes to database")
	status.Set("current", 0)
	s.app.Save(status)

	var i int
	for _, match := range result.Matches {
		record, err := s.app.FindFirstRecordByData("games", "path", match.Path)
		if err != nil {
			record = core.NewRecord(s.gamesCollection)
		}

		record.Set("provider", match.Game.Provider)
		record.Set("providerId", match.Game.ProviderID)

		record.Set("name", match.Game.Name)
		record.Set("path", match.Path)
		record.Set("summary", match.Game.Summary)
		record.Set("status", "found")
		record.Set("version", match.FilenameMetadata.Version)
		record.Set("released", match.Game.ReleaseDate)
		record.Set("rating", match.Game.Rating)

		marshaledGenres, err := json.Marshal(match.Game.Genres)
		if err != nil {
			return err
		}
		if slices.Equal(marshaledGenres, []byte("null")) {
			marshaledGenres = []byte("[]")
		}
		record.Set("genres", string(marshaledGenres))

		cover, err := filesystem.NewFileFromURL(ctx, match.Game.Cover)
		if err != nil {
			return err
		}
		record.Set("cover", cover)

		var artworks []*filesystem.File
		for _, url := range match.Game.Artworks {
			artwork, err := filesystem.NewFileFromURL(ctx, url)
			if err != nil {
				return err
			}
			artworks = append(artworks, artwork)
		}
		record.Set("artworks", artworks)

		var screenshots []*filesystem.File
		for _, url := range match.Game.Screenshots {
			screenshot, err := filesystem.NewFileFromURL(ctx, url)
			if err != nil {
				return err
			}
			screenshots = append(screenshots, screenshot)
		}
		record.Set("screenshots", screenshots)

		if err := s.app.Save(record); err != nil {
			return err
		}

		i++
		status.Set("current", i)
		s.app.Save(status)
	}
	for _, missing := range result.Missing {
		record, err := s.app.FindFirstRecordByData("games", "path", missing.Path)
		if err != nil {
			record = core.NewRecord(s.gamesCollection)
		}

		record.Set("path", missing.Path)
		record.Set("name", missing.FilenameMetadata.Name)
		record.Set("version", missing.FilenameMetadata.Version)
		record.Set("status", "missing")

		if err := s.app.Save(record); err != nil {
			return err
		}

		i++
		status.Set("current", i)
		s.app.Save(status)
	}
	for _, path := range result.Invalid {
		record, err := s.app.FindFirstRecordByData("games", "path", path)
		if err != nil {
			record = core.NewRecord(s.gamesCollection)
		}

		record.Set("path", path)
		record.Set("name", path)
		record.Set("status", "invalid")

		if err := s.app.Save(record); err != nil {
			return err
		}

		i++
		status.Set("current", i)
		s.app.Save(status)
	}

	return nil
}
