package main

import (
	"embed"
	"net/http"
	"os"
	"runtime"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"

	"github.com/moniquelive/tv-gin/handler"
	"github.com/moniquelive/tv-gin/router"
)

//go:embed web
var embedFS embed.FS

//go:embed web/cfg/config.json
var configFile []byte

func init() {
	handler.Init(embedFS, configFile)
}

func main() {
	var engine *html.Engine
	if runtime.GOOS == "linux" {
		engine = productionEngine()
	} else {
		engine = developmentEngine()
	}
	//engine.AddFuncMap(sprig.HtmlFuncMap())
	fiberConfig := fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	}
	corsConfig := cors.Config{
		AllowOrigins: os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowHeaders: "Origin, Content-Type, Accept",
	}
	app := fiber.New(fiberConfig)
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

func developmentEngine() (engine *html.Engine) {
	engine = html.NewFileSystem(http.FS(os.DirFS("web")), ".html")
	engine.Reload(true)
	engine.Debug(true)
	return
}

func productionEngine() (engine *html.Engine) {
	engine = html.NewFileSystem(http.FS(embedFS), ".html")
	engine.Directory = "web"
	return
}
