package download

import (
	"boyl/client/pkg/archive"
	"boyl/client/pkg/remote"
	"boyl/client/pkg/settings"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/pocketbase/pocketbase/core"
)

type Download struct {
	record         *core.Record
	app            core.App
	settings       *settings.Settings
	remote         *remote.Client
	game           *remote.Game
	gamesDirectory string
	baseDirectory  string
	archivePath    string
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewDownload(record *core.Record, app core.App, settings *settings.Settings, remote *remote.Client) (*Download, error) {
	gamesDirectory := settings.GetString("gamesDirectory")

	game, err := remote.GetGame(record.GetString("game"))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Download{
		record:         record,
		app:            app,
		settings:       settings,
		remote:         remote,
		game:           game,
		gamesDirectory: gamesDirectory,
		baseDirectory:  filepath.Join(gamesDirectory, game.Name),
		archivePath:    filepath.Join(gamesDirectory, game.Name, game.ID+".tmp"),
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

func (d *Download) Start() error {
	status := d.record.GetString("status")
	if status == "starting" {
		d.record.Set("status", "downloading")
		d.record.Set("progress", 0)
		d.record.Set("speed", 0)
		d.record.Set("total", 0)
		err := d.app.Save(d.record)
		if err != nil {
			return err
		}

		err = d.Start()
		if err != nil {
			return err
		}
	}
	if status == "downloading" {
		err := d.download()
		if err != nil {
			d.record.Set("status", "failed")
			d.record.Set("text", err.Error())
			d.app.Save(d.record)
			return err
		}

		d.record.Set("status", "extracting")
		d.record.Set("progress", 0)
		d.record.Set("speed", 0)
		d.record.Set("total", 0)
		err = d.app.Save(d.record)
		if err != nil {
			return err
		}

		err = d.Start()
		if err != nil {
			return err
		}
	}
	if status == "extracting" {
		err := d.extract()
		if err != nil {
			d.record.Set("status", "failed")
			d.record.Set("text", err.Error())
			d.app.Save(d.record)
			return err
		}

		d.record.Set("status", "completed")
		d.record.Set("progress", 1)
		err = d.app.Save(d.record)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Download) download() error {
	d.record.Set("status", "downloading")
	d.app.Save(d.record)

	client := grab.NewClient()
	client.HTTPClient = d.remote.Client()
	req, err := grab.NewRequest(
		d.archivePath,
		fmt.Sprintf("%s/api/download?id=%s", d.remote.URL, d.game.ID),
	)
	if err != nil {
		return err
	}

	resp := client.Do(req)
	d.record.Set("total", resp.HTTPResponse.ContentLength)
	d.app.Save(d.record)

	t := time.NewTicker(500 * time.Millisecond)

loop:
	for {
		select {
		case <-t.C:
			d.record.Set("progress", resp.Progress())
			d.record.Set("speed", resp.BytesPerSecond())
			d.app.Save(d.record)
		case <-d.ctx.Done():
			return resp.Cancel()
		case <-resp.Done:
			break loop
		}
	}

	if err := resp.Err(); err != nil {
		return err
	}

	return nil
}

func (d *Download) extract() error {
	d.record.Set("status", "extracting")
	d.app.Save(d.record)

	file, err := os.Open(d.archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	size := info.Size()

	extractor, err := archive.NewExtractor(filepath.Base(d.game.Path), file, size)
	if err != nil {
		return err
	}

	progressSize, err := extractor.GetProgressSize()
	if err != nil {
		return err
	}
	d.record.Set("total", progressSize)
	d.app.Save(d.record)

	movingAverage := NewMovingAverage(5 * time.Second)
	lastProgress := time.Now()
	var lastValue uint64

	err = extractor.Extract(d.ctx, d.baseDirectory, func(u uint64) {
		if time.Since(lastProgress) > 500*time.Millisecond {
			movingAverage.Add(float64(u - lastValue))
			lastValue = u
			d.record.Set("progress", float64(u)/float64(progressSize))
			d.record.Set("speed", movingAverage.Get())
			d.app.Save(d.record)
			lastProgress = time.Now()
		}
	})
	if err != nil {
		return err
	}

	if err := os.Remove(d.archivePath); err != nil {
		return err
	}

	return nil
}
