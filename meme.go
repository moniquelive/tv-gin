package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

const (
	dpi          = 72
	fontFileName = "./static/jetbrains.ttf"
	size         = 64
	spacing      = 1
)

var ttfont *truetype.Font

func init() {
	fontBytes, err := ioutil.ReadFile(fontFileName)
	if err != nil {
		log.Fatalln(err)
	}
	ttfont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		log.Fatalln(err)
	}
}

func generateMeme(memeFilename string, texts [2]string, rects [2]image.Rectangle, leftMargin int) (*bytes.Buffer, error) {
	img, err := getImageFromFilePath(memeFilename)
	if err != nil {
		return nil, fmt.Errorf("getImageFromFilePath: %w", err)
	}

	canvas := image.NewRGBA(img.Bounds())
	draw.Draw(canvas, canvas.Bounds(), img, img.Bounds().Min, draw.Src)
	for x := 0; x < img.Bounds().Max.X; x++ {
		canvas.Set(x, img.Bounds().Max.Y/2, color.Black)
	}

	fg, _ := image.Black, image.White
	fc := freetype.NewContext()
	fc.SetDPI(dpi)
	fc.SetFont(ttfont)
	fc.SetFontSize(size)
	fc.SetClip(canvas.Bounds())
	fc.SetDst(canvas)
	fc.SetSrc(fg)
	fc.SetHinting(font.HintingNone)
	//fc.SetHinting(font.HintingFull)

	const wordwrapWidth = 13

	// Draw the text.
	y := measureString(fc, wordWrap(texts[0], wordwrapWidth))
	rect0HalfHeight := rects[0].Dy() / 2
	err = drawString(fc,
		wordWrap(texts[0], wordwrapWidth),
		rects[0].Min.X+leftMargin,
		rects[0].Min.Y+rect0HalfHeight-y/2)
	if err != nil {
		return nil, fmt.Errorf("drawString (1): %w", err)
	}

	y = measureString(fc, wordWrap(texts[1], wordwrapWidth))
	rect1HalfHeight := rects[1].Dy() / 2
	err = drawString(fc,
		wordWrap(texts[1], wordwrapWidth),
		rects[1].Min.X+leftMargin,
		rects[1].Min.Y+rect1HalfHeight-y/2)
	if err != nil {
		return nil, fmt.Errorf("drawString (2): %w", err)
	}

	opts := jpeg.Options{Quality: 99}
	rw := bytes.Buffer{}
	err = jpeg.Encode(&rw, canvas, &opts)
	if err != nil {
		return nil, fmt.Errorf("jpeg.Encode: %w", err)
	}

	return &rw, nil
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
