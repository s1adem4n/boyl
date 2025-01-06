package download

import (
	"boyl/client/pkg/remote"
	"boyl/client/pkg/settings"
	"fmt"
	"sync"

	"github.com/pocketbase/pocketbase/core"
)

type Manager struct {
	app                 core.App
	downloadsCollection *core.Collection
	gamesCollection     *core.Collection
	settings            *settings.Settings
	remote              *remote.Client

	mu        sync.Mutex
	downloads map[string]*Download
}

func NewManager(app core.App, downloadsCollection *core.Collection, gamesCollection *core.Collection, settings *settings.Settings, remote *remote.Client) *Manager {
	return &Manager{
		app:                 app,
		downloadsCollection: downloadsCollection,
		gamesCollection:     gamesCollection,
		settings:            settings,
		remote:              remote,
		mu:                  sync.Mutex{},
		downloads:           make(map[string]*Download),
	}
}

func (m *Manager) Worker(records chan *core.Record) {
	for record := range records {
		if record == nil {
			m.app.Logger().Error("nil record")
			continue
		}
		download, err := NewDownload(record, m.app, m.settings, m.remote)
		if err != nil {
			m.app.Logger().Error("failed to create download", "error", err)
			continue
		}

		m.mu.Lock()
		m.downloads[record.Id] = download
		m.mu.Unlock()

		err = download.Start()
		if err != nil {
			m.app.Logger().Error("failed to start download", "error", err)
			continue
		}

		game, err := m.app.FindFirstRecordByData(m.gamesCollection, "game", record.GetString("game"))
		if err != nil {
			game = core.NewRecord(m.gamesCollection)
		}
		game.Set("game", download.game.ID)
		game.Set("path", download.baseDirectory)

		executable := download.game.Executable
		if executable == "" {
			executable, err = FindExecutablePath(download.baseDirectory)
			if err != nil {
				m.app.Logger().Error("failed to find executable", "error", err)
				continue
			}
		}
		game.Set("executable", executable)

		err = m.app.Save(game)
		if err != nil {
			m.app.Logger().Error("failed to save game", "error", err)
			continue
		}
	}
}

func (m *Manager) Cancel(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	download, ok := m.downloads[id]
	if !ok {
		return fmt.Errorf("download %s not found", id)
	}

	status := download.record.GetString("status")
	if status == "completed" || status == "failed" {
		delete(m.downloads, id)
		return m.app.Delete(download.record)
	}

	download.cancel()
	download.record.Set("status", "failed")
	return m.app.Save(download.record)
}
