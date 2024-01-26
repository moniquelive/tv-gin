package handler

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"

	"github.com/moniquelive/tv-gin/meme"
)

var memes *meme.Memes

var embedFS embed.FS

func Init(embed embed.FS, configFile []byte) {
	embedFS = embed

	var err error
	memes, err = meme.NewMeme(configFile)
	if err != nil {
		log.Fatalf("NewMeme: %v", err)
	}
}

func Meme(c *fiber.Ctx) error {
	var textMessages struct {
		Text []string
	}
	if err := c.QueryParser(&textMessages); err != nil {
		return fiber.NewError(http.StatusInternalServerError, fmt.Sprintf("QueryParser: %v", err))
	}
	memeName := c.Query("meme")
	theMeme, err := memes.FindMeme(memeName)
	if err != nil {
		return fiber.NewError(http.StatusNotFound, fmt.Sprintf("FindMeme: %v", err))
	}
	memeImage, err := theMeme.Generate(embedFS, textMessages.Text)
	if err != nil {
		return fiber.NewError(http.StatusInternalServerError, fmt.Sprintf("Generate: %v", err))
	}
	c.Set(fiber.HeaderContentType, "image/jpeg")
	return c.SendStream(memeImage, memeImage.Len())
}
