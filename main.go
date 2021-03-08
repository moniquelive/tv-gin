package main

import (
	"bytes"
	"github.com/golang/freetype/truetype"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang/freetype"
	"golang.org/x/image/font"
)

var r = gin.Default()
var ttfont *truetype.Font

const (
	dpi      = 72
	fontfile = "jetbrains.ttf"
	size     = 64
	spacing  = 1.5
)

func init() {
	r.GET("/ping", pingHandler)
	r.GET("/meme", memeHandler)

	fontBytes, err := ioutil.ReadFile(fontfile)
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
	//fc.SetHinting(font.HintingNone)
	fc.SetHinting(font.HintingFull)

	// Draw the text.
	err = drawString(fc, []string{text1}, 655, 40)
	if err != nil {
		return
	}

	err = drawString(fc, []string{text2}, 655, 650)
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

func getImageFromFilePath(filePath string) (image.Image, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func main() {
	if err := r.Run(); err != nil {
		log.Fatalln("r.Run:", err)
	}
}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
