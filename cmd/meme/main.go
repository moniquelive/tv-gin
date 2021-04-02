package main

import (
	"github.com/moniquelive/tv-gin/internal/meme"
	"github.com/moniquelive/tv-gin/internal/utils"

	"embed"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

//go:embed static
var f embed.FS

var r = gin.Default()

func init() {
	r.Use(static.Serve("/", utils.EmbedFolder(f, "static")))
	r.GET("/meme", memeHandler)
}

func main() {
	if err := r.Run(); err != nil {
		log.Fatalln("r.Run:", err)
	}
}

func memeHandler(c *gin.Context) {
	const param1Name = "text1"
	const param2Name = "text2"
	texts := []string{
		c.Query(param1Name),
		c.Query(param2Name),
	}
	if texts[0] == "" && texts[1] == "" {
		c.String(http.StatusBadRequest, fmt.Sprintf(`parâmetros %q e %q são obrigatórios`, param1Name, param2Name))
		return
	}
	if texts[0] == "" {
		c.String(http.StatusBadRequest, fmt.Sprintf(`parâmetro %q é obrigatório`, param1Name))
		return
	}
	if texts[1] == "" {
		c.String(http.StatusBadRequest, fmt.Sprintf(`parâmetro %q é obrigatório`, param2Name))
		return
	}

	buffer, err := meme.Generate("drake", texts)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("generateMeme> %v", err))
		return
	}
	c.DataFromReader(http.StatusOK, int64(buffer.Len()), "image/jpeg", buffer, nil)
}
