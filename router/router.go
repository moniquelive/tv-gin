package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/monitor"

	"github.com/moniquelive/tv-gin/handler"
)

func SetupRoutes(app *fiber.App) {
	app.Static("/css", "./web/css").
		Static("/img", "./web/img").
		Static("/js", "./web/js").
		Static("/cfg", "./web/cfg")

	app.Get("/status", monitor.New()).
		Get("/", handler.Index).
		Get("/meme", handler.Meme)
}
