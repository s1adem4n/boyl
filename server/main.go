package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "boyl/server/migrations"
	"boyl/server/scan"
	"boyl/server/scan/metadata"
	"boyl/server/scan/metadata/gog"
	"boyl/server/scan/metadata/igdb"
	"boyl/server/scan/metadata/steam"
)

func loadEnv(name string) string {
	value := os.Getenv(name)
	if value == "" {
		log.Fatalf("%s environment variable is not set", name)
	}
	return value
}

func main() {
	godotenv.Load()

	app := pocketbase.New()

	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		Automigrate: isGoRun,
	})

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.InstallerFunc = func(app core.App, systemSuperuser *core.Record, baseURL string) error {
			email := os.Getenv("ADMIN_EMAIL")
			password := os.Getenv("ADMIN_PASSWORD")
			if email == "" || password == "" {
				return errors.New("ADMIN_EMAIL and ADMIN_PASSWORD environment variables are not set, admin user will not be created")
			}
			systemSuperuser.SetEmail(email)
			systemSuperuser.SetPassword(password)

			return app.Save(systemSuperuser)
		}

		gamesDirectory := loadEnv("GAMES_DIRECTORY")
		igdbClientID := loadEnv("IGDB_CLIENT_ID")
		igdbClientSecret := loadEnv("IGDB_CLIENT_SECRET")

		gamesCollection, err := app.FindCollectionByNameOrId("games")
		if err != nil {
			return err
		}
		statusCollection, err := app.FindCollectionByNameOrId("status")
		if err != nil {
			return err
		}

		igdbProvider := igdb.NewProvider(igdbClientID, igdbClientSecret)
		gogProvider := gog.NewProvider()
		steamProvider := steam.NewProvider()
		scanner := scan.NewScanner(
			gamesDirectory,
			[]metadata.Provider{
				igdbProvider,
				gogProvider,
				steamProvider,
			},
			app,
			gamesCollection,
			statusCollection,
		)

		se.Router.GET("/api/download", func(e *core.RequestEvent) error {
			if e.Auth == nil || (!e.Auth.IsSuperuser() && e.Auth.GetBool("verified") != true) {
				return e.UnauthorizedError("unauthorized", nil)
			}

			id := e.Request.URL.Query().Get("id")
			if id == "" {
				return e.BadRequestError("id is required", nil)
			}
			game, err := app.FindRecordById("games", id)
			if err != nil {
				return e.BadRequestError("game not found", nil)
			}

			path := game.GetString("path")
			status := game.GetString("status")
			if path == "" || status == "deleted" {
				return e.BadRequestError("game is not available for download", nil)
			}

			file, err := os.Open(path)
			if err != nil {
				return e.InternalServerError("error while opening file", err)
			}
			defer file.Close()

			info, err := file.Stat()
			if err != nil {
				return e.InternalServerError("error while getting file info", err)
			}

			log.Printf("kb size is %d", info.Size()/1024)

			http.ServeContent(e.Response, e.Request, info.Name(), info.ModTime(), file)
			return nil
		})

		se.Router.GET("/api/scan", func(e *core.RequestEvent) error {
			if e.Auth == nil || (!e.Auth.IsSuperuser() && e.Auth.GetBool("verified") != true) {
				return e.UnauthorizedError("unauthorized", nil)
			}

			if scanner.IsScanning() {
				return e.JSON(200, "Scanning in progress")
			}

			go func() {
				err := scanner.Update(context.Background())
				if err != nil {
					app.Logger().Error("error while scanning", "error", err)
				}
			}()

			return nil
		})
		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
