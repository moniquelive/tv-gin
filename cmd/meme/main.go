package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/moniquelive/tv-gin/internal/meme"

	"github.com/gin-gonic/gin"
)

//go:embed web
var webFS embed.FS

func main() {
	r := gin.Default()
	r.GET("/*resource", func(c *gin.Context) {
		resourceName := c.Param("resource")
		switch resourceName {
		case "/meme":
			serveMeme(c)
		default:
			c.FileFromFS("/web"+resourceName, http.FS(webFS))
		}
	})
	if err := r.Run(); err != nil {
		log.Fatalln("r.Run:", err)
	}
}

func serveMeme(c *gin.Context) {
	memeName := c.Query("meme")
	textMessages := c.QueryArray("text[]")
	theMeme, err := meme.NewMeme(webFS).FindMeme(memeName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("FindMeme> %v", err))
		return
	}
	memeImage, err := theMeme.Generate(textMessages)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("generateMeme> %v", err))
		return
	}
	c.DataFromReader(http.StatusOK, int64(memeImage.Len()), "image/jpeg", memeImage, nil)
}
