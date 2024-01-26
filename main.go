package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/moniquelive/tv-gin/handler"
	"github.com/moniquelive/tv-gin/router"
)

//go:embed web
var embedFS embed.FS

//go:embed web/cfg/config.json
var configFile []byte

func init() {
	handler.Init(embedFS, configFile)
	router.Init(embedFS)
}

func main() {
	corsConfig := cors.Config{
		AllowOrigins: os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowHeaders: "Origin, Content-Type, Accept",
	}
	app := fiber.New()
	app.Use(favicon.New()).
		Use(logger.New()).
		Use(helmet.New()).
		Use(recover.New()).
		Use(cors.New(corsConfig))
	router.SetupRoutes(app)

	if runtime.GOOS == "linux" {
		log.Fatal(app.Listen(":8080"))
	} else {
		log.Fatal(app.Listen("localhost:8080"))
	}
}
