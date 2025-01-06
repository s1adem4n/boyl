package main

import (
	"boyl/client/frontend"
	"boyl/client/pkg/download"
	"boyl/client/pkg/remote"
	"boyl/client/pkg/settings"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/google/shlex"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	webview "github.com/webview/webview_go"

	_ "boyl/client/migrations"
)

func getRemote(s *settings.Settings) (*remote.Client, error) {
	serverURL := s.GetString("serverUrl")
	email := s.GetString("email")
	password := s.GetString("password")
	r := remote.New(serverURL)

	if email != "" && password != "" {
		err := r.Authenticate(email, password)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func main() {
	app := pocketbase.New()

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	downloadsChannel := make(chan *core.Record, 20)

	app.OnRecordCreate("downloads").BindFunc(func(e *core.RecordEvent) error {
		downloadsChannel <- e.Record

		return e.Next()
	})

	app.OnTerminate().BindFunc(func(te *core.TerminateEvent) error {
		downloadsCollection, err := app.FindCollectionByNameOrId("downloads")
		if err != nil {
			return err
		}
		downloads, err := app.FindAllRecords(downloadsCollection)
		if err != nil {
			return err
		}
		for _, download := range downloads {
			download.Set("active", false)
			if err := app.Save(download); err != nil {
				return err
			}
		}

		return te.Next()
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.InstallerFunc = func(app core.App, systemSuperuser *core.Record, baseURL string) error {
			systemSuperuser.SetEmail("admin@local.host")
			systemSuperuser.SetPassword("boyladmin")
			return app.Save(systemSuperuser)
		}

		subFS := apis.MustSubFS(frontend.Assets, "build")
		se.Router.GET("/{path...}", apis.Static(subFS, true))

		downloadsCollection, err := app.FindCollectionByNameOrId("downloads")
		if err != nil {
			return err
		}
		settingsCollection, err := app.FindCollectionByNameOrId("settings")
		if err != nil {
			return err
		}
		gamesCollection, err := app.FindCollectionByNameOrId("games")
		if err != nil {
			return err
		}

		r, err := getRemote(settings.NewSettings(app, settingsCollection))
		if err != nil {
			return err
		}

		s := settings.NewSettings(app, settingsCollection)
		m := download.NewManager(app, downloadsCollection, gamesCollection, s, r)
		go m.Worker(downloadsChannel)

		downloads, err := app.FindAllRecords(downloadsCollection)
		if err != nil {
			return err
		}
		for _, download := range downloads {
			downloadsChannel <- download
		}

		if err := s.Set("os", runtime.GOOS); err != nil {
			return err
		}

		se.Router.GET("/api/update-remote", func(e *core.RequestEvent) error {
			serverURL := s.GetString("serverUrl")
			email := s.GetString("email")
			password := s.GetString("password")

			r.URL = serverURL
			if email != "" && password != "" {
				err := r.Authenticate(email, password)
				if err != nil {
					return err
				}
			}

			return e.JSON(200, "")
		})

		se.Router.POST("/api/launch", func(e *core.RequestEvent) error {
			q := e.Request.URL.Query()
			id := q.Get("id")
			if id == "" {
				return e.BadRequestError("id is required", nil)
			}

			game, err := app.FindRecordById(gamesCollection, id)
			if err != nil {
				return e.NotFoundError("game not found", nil)
			}

			executable := game.GetString("executable")
			if executable == "" {
				return e.BadRequestError("executable is required", nil)
			}

			userArgs, err := shlex.Split(game.GetString("args"))
			if err != nil {
				return e.BadRequestError("failed to parse args", err)
			}

			launcherString := game.GetString("launcher")
			if launcherString == "" {
				launcherString = s.GetString("defaultLauncher")
			}
			launcher, err := shlex.Split(launcherString)
			if err != nil {
				return e.BadRequestError("failed to parse launcher", err)
			}

			var name string
			var args []string

			if len(launcher) > 0 {
				name = launcher[0]
				args = append(launcher[1:], executable)
			} else {
				name = executable
			}
			args = append(args, userArgs...)

			cmd := exec.Command(name, args...)
			cmd.Env = os.Environ()
			if err := cmd.Start(); err != nil {
				return e.InternalServerError("failed to start", err)
			}
			cmd.Process.Release()

			return e.JSON(200, "")
		})

		se.Router.POST("/api/download", func(e *core.RequestEvent) error {
			q := e.Request.URL.Query()
			id := q.Get("id")
			if id == "" {
				return e.BadRequestError("id is required", nil)
			}

			download := core.NewRecord(downloadsCollection)
			download.Set("game", id)
			download.Set("status", "starting")
			if err := app.Save(download); err != nil {
				return err
			}

			return e.JSON(200, "")
		})

		se.Router.DELETE("/api/download", func(e *core.RequestEvent) error {
			q := e.Request.URL.Query()
			id := q.Get("id")
			if id == "" {
				return e.BadRequestError("id is required", nil)
			}

			if err := m.Cancel(id); err != nil {
				return err
			}

			return e.JSON(200, "")
		})

		return se.Next()
	})

	if len(os.Args) == 1 {
		go func() {
			w := webview.New(true)

			w.SetTitle("Boyl")
			w.SetSize(480, 320, webview.HintMin)
			w.Navigate("http://localhost:48658")
			w.Run()
			w.Destroy()
			os.Exit(0)
		}()

		err := app.Bootstrap()
		if err != nil {
			log.Fatal(err)
		}
		err = apis.Serve(app, apis.ServeConfig{
			HttpAddr:        "localhost:48658",
			ShowStartBanner: false,
		})
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	} else {
		err := app.Start()
		if err != nil {
			log.Fatal(err)
		}
	}
}
