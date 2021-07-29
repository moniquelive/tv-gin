package main

import (
	"github.com/gin-contrib/static"
	"github.com/moniquelive/tv-gin/internal/meme"

	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)


func main() {
	var r = gin.Default()
	r.Use(static.Serve("/", static.LocalFile("web", false)))
	r.GET("/config.json", func(c *gin.Context) {
		c.JSON(http.StatusOK, meme.Memes)
	})
	r.GET("/meme", memeHandler)
	if err := r.Run(); err != nil {
		log.Fatalln("r.Run:", err)
	}
}

func memeHandler(c *gin.Context) {
	const paramMeme = "meme"
	const paramTexts = "text[]"
	memeName := c.Query(paramMeme)
	textsVal := c.QueryArray(paramTexts)
	theMeme, err := meme.Memes.FindMeme(memeName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("FindMeme> %v", err))
		return
	}
	//if len(textsVal) != theMeme.NumBoxes() {
	//	c.String(http.StatusBadRequest,
	//		fmt.Sprintf(`Precisamos de %d linhas de texto! %d recebida(s)...`,
	//			theMeme.NumBoxes(),
	//			len(textsVal)))
	//	return
	//}
	buffer, err := theMeme.Generate(textsVal)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("generateMeme> %v", err))
		return
	}
	c.DataFromReader(http.StatusOK, int64(buffer.Len()), "image/jpeg", buffer, nil)
}
