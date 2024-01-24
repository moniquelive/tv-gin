package main

import (
	"embed"
	"fmt"
	"github.com/moniquelive/tv-gin/meme"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed web
var webRoot embed.FS

//go:embed web/config.json
var configFile []byte

var memes *meme.Memes

func init() {
	var err error
	memes, err = meme.NewMeme(configFile)
	if err != nil {
		log.Fatalf("NewMeme: %v", err)
	}
}

func serveMeme(c *gin.Context) {
	memeName := c.Query("meme")
	textMessages := c.QueryArray("text[]")
	theMeme, err := memes.FindMeme(memeName)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, fmt.Errorf("FindMeme: %w", err))
		return
	}
	memeImage, err := theMeme.Generate(webRoot, textMessages)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("generateMeme: %w", err))
		return
	}
	c.DataFromReader(http.StatusOK, int64(memeImage.Len()), "image/jpeg", memeImage, nil)
}

func main() {
	r := gin.Default()
	r.Use(gin.ErrorLogger())
	r.GET("/*resource", func(c *gin.Context) {
		resourceName := c.Param("resource")
		switch resourceName {
		case "/meme":
			serveMeme(c)
		default:
			c.FileFromFS("/web"+resourceName, http.FS(webRoot))
		}
	})
	if err := r.Run(); err != nil {
		log.Fatalln("r.Run:", err)
	}
}
