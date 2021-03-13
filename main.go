package main

import (
	"embed"
	"image"
	"log"
	"net/http"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

//go:embed static
var f embed.FS

//go:embed static/meme.jpg
var meme []byte

var r = gin.Default()

func init() {
	r.Use(static.Serve("/", EmbedFolder(f, "static")))
	r.GET("/meme", memeHandler)
}

func main() {
	if err := r.Run(); err != nil {
		log.Fatalln("r.Run:", err)
	}
}

func memeHandler(c *gin.Context) {
	text1 := c.Query("text1")
	text2 := c.Query("text2")
	if text1 == "" || text2 == "" {
		c.String(http.StatusBadRequest, `parâmetros "text1" e "text2" são obrigatórios`)
		return
	}
	texts := [2]string{text1, text2}
	rects := [2]image.Rectangle{
		{image.Point{X: 600, Y: 0}, image.Point{X: 1199, Y: 599}},
		{image.Point{X: 600, Y: 600}, image.Point{X: 1199, Y: 1199}},
	}
	margin := 55
	buffer, err := generateMeme(meme, texts, rects, margin)
	if err != nil {
		c.String(http.StatusInternalServerError, "generateMeme:", err)
		return
	}
	c.DataFromReader(http.StatusOK, int64(buffer.Len()), "image/jpeg", buffer, nil)
}
