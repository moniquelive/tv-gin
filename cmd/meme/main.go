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
	var r = gin.Default()
	r.GET("/*page", func(c *gin.Context) {
		pageName := c.Param("page")
		if pageName == "/meme" {
			serveMeme(c)
			return
		}
		c.FileFromFS("/web"+pageName, http.FS(webFS))
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
	buffer, err := theMeme.Generate(textMessages)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("generateMeme> %v", err))
		return
	}
	c.DataFromReader(http.StatusOK, int64(buffer.Len()), "image/jpeg", buffer, nil)
}
