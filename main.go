package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

var r = gin.Default()
var ttfont *truetype.Font

const (
	dpi      = 72
	fontFile = "jetbrains.ttf"
	size     = 64
	spacing  = 1
)

func init() {
	r.StaticFile("/", "./static/index.html")
	r.GET("/ping", pingHandler)
	r.GET("/meme", memeHandler)

	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		log.Println(err)
		return
	}
	ttfont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		log.Println(err)
		return
	}
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
	img, err := getImageFromFilePath("meme.jpg")
	if err != nil {
		c.String(http.StatusInternalServerError, "getImageFromFilePath:", err)
		return
	}

	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	for x := 0; x < img.Bounds().Max.X; x++ {
		rgba.Set(x, img.Bounds().Max.Y/2, color.Black)
	}

	fg, _ := image.Black, image.White
	fc := freetype.NewContext()
	fc.SetDPI(dpi)
	fc.SetFont(ttfont)
	fc.SetFontSize(size)
	fc.SetClip(rgba.Bounds())
	fc.SetDst(rgba)
	fc.SetSrc(fg)
	fc.SetHinting(font.HintingNone)
	//fc.SetHinting(font.HintingFull)

	// Draw the text.
	y := measureString(fc, wordWrap(text1, 13))
	err = drawString(fc, wordWrap(text1, 13), 655, 300-y/2)
	if err != nil {
		return
	}

	y = measureString(fc, wordWrap(text2, 13))
	err = drawString(fc, wordWrap(text2, 13), 655, 600+300-y/2)
	if err != nil {
		return
	}

	opts := jpeg.Options{Quality: 99}
	rw := bytes.Buffer{}
	err = jpeg.Encode(&rw, rgba, &opts)
	if err != nil {
		c.String(http.StatusInternalServerError, "jpeg.Encode:", err)
		return
	}
	c.DataFromReader(http.StatusOK, int64(rw.Len()), "image/jpeg", &rw, nil)
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func drawString(fc *freetype.Context, text []string, x, y int) error {
	pt := freetype.Pt(x, y+int(fc.PointToFixed(size)>>6))
	for _, s := range text {
		_, err := fc.DrawString(s, pt)
		if err != nil {
			log.Println(err)
			return err
		}
		pt.Y += fc.PointToFixed(size * spacing)
	}
	return nil
}

func measureString(fc *freetype.Context, text []string) int {
	return int(fc.PointToFixed(size)>>6) +
		(len(text)-1)*(int(fc.PointToFixed(size*spacing))>>6)
}

func wordWrap(text string, lineWidth int) (lines []string) {
	words := strings.Fields(strings.TrimSpace(text))
	if len(words) == 0 {
		return []string{}
	}
	lines = []string{words[0]}
	spaceLeft := lineWidth - len(lines[len(lines)-1])
	for _, word := range words[1:] {
		if len(word)+1 > spaceLeft {
			lines = append(lines, word)
			spaceLeft = lineWidth - len(word)
		} else {
			lines[len(lines)-1] += " " + word
			spaceLeft -= 1 + len(word)
		}
	}
	return
}
