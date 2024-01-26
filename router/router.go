package router

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/monitor"

	"github.com/moniquelive/tv-gin/handler"
)

var embedFS embed.FS

func Init(embed embed.FS) {
	embedFS = embed
}

func SetupRoutes(app *fiber.App) {
	app.Get("/status", monitor.New()).
		Get("/meme", handler.Meme)

	app.Use(filesystem.New(filesystem.Config{
		Root:       http.FS(embedFS),
		PathPrefix: "web",
	}))
}
