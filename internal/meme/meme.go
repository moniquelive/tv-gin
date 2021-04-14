package meme

import (
	"github.com/moniquelive/tv-gin/internal/utils"

	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

//go:embed jetbrains.ttf
var fontBytes []byte

//go:embed Handlee-Regular.ttf
var creditsFontBytes []byte

//go:embed config.json
var configJson []byte

var memeFont *truetype.Font
var creditsFont *truetype.Font

type meme struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Filename   string   `json:"filename"`
	FontColor  string   `json:"font-color"`
	LineChars  int      `json:"line-chars"`
	MarginLeft int      `json:"margin-left"`
	TextAlign  string   `json:"text-align"`
	Boxes      [][4]int `json:"boxes"`
}

type config struct {
	Memes []meme `json:"memes"`
}

var Config config

func init() {
	var err error
	memeFont, err = freetype.ParseFont(fontBytes)
	if err != nil {
		log.Fatalln(err)
	}
	creditsFont, err = freetype.ParseFont(creditsFontBytes)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(configJson, &Config)
	if err != nil {
		log.Fatalln(err)
	}
}

func (c config) FindMeme(name string) (*meme, error) {
	if name == "" {
		name = Config.Memes[0].ID
	}
	for _, meme := range c.Memes {
		if meme.ID == name {
			return &meme, nil
		}
	}
	return nil, fmt.Errorf("meme nao encontrado: %q", name)
}

func (m meme) NumBoxes() int {
	return len(m.Boxes)
}

func (m meme) FontRGBA() color.RGBA {
	c, err := utils.ParseHexColor(m.FontColor)
	if err != nil {
		log.Panicf("FontRGBA> %v", err)
	}
	return c
}

func (m meme) Generate(texts []string) (*bytes.Buffer, error) {
	coords := m.Boxes
	rects := [2]image.Rectangle{
		{image.Point{X: coords[0][0], Y: coords[0][1]}, image.Point{X: coords[0][2], Y: coords[0][3]}},
		{image.Point{X: coords[1][0], Y: coords[1][1]}, image.Point{X: coords[1][2], Y: coords[1][3]}},
	}
	f, err := os.Open("./web/" + m.Filename)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("image.Decode: %w", err)
	}

	canvas := image.NewRGBA(img.Bounds())
	draw.Draw(canvas, canvas.Bounds(), img, img.Bounds().Min, draw.Src)

	// DEBUG!
	//red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	//for _, r := range m.Boxes {
	//	utils.Rect(canvas, red, r[0], r[1], r[2], r[3])
	//}

	const (
		memeFontSize    = 64
		memeFontSpacing = 1
	)
	fc := utils.CreateFontContext(memeFont, memeFontSize, canvas.Bounds(), canvas, m.FontRGBA())

	// Draw the text.
	wrap := wordWrap(texts[0], m.LineChars)
	y := utils.TextHeight(fc, memeFontSize, memeFontSpacing, wrap)
	rect0HalfHeight := rects[0].Dy() / 2
	err = utils.DrawString(fc,
		memeFont, memeFontSize, rects[0],
		m.TextAlign,
		memeFontSpacing,
		wrap,
		rects[0].Min.X+m.MarginLeft,
		rects[0].Min.Y+rect0HalfHeight-y/2)
	if err != nil {
		return nil, fmt.Errorf("DrawString (1): %w", err)
	}

	wrap = wordWrap(texts[1], m.LineChars)
	y = utils.TextHeight(fc, memeFontSize, memeFontSpacing, wrap)
	rect1HalfHeight := rects[1].Dy() / 2
	err = utils.DrawString(fc,
		memeFont, memeFontSize, rects[1],
		m.TextAlign,
		memeFontSpacing,
		wrap,
		rects[1].Min.X+m.MarginLeft,
		rects[1].Min.Y+rect1HalfHeight-y/2)
	if err != nil {
		return nil, fmt.Errorf("DrawString (2): %w", err)
	}

	const (
		creditsFontSize = 32
		creditsText     = "Esta imagem foi gerada no meme.monique.dev"
	)
	creditsFontColor := color.RGBA{R: 0xfe, G: 0x43, B: 0x65, A: 0xff}
	fc = utils.CreateFontContext(creditsFont, creditsFontSize, canvas.Bounds(), canvas, creditsFontColor)
	var (
		creditsWidth      = utils.TextWidthInPixels(creditsFont, creditsFontSize, creditsText)
		creditsX          = canvas.Bounds().Max.X - creditsWidth - creditsFontSize
		creditsFontHeight = int(fc.PointToFixed(creditsFontSize) >> 6)
		creditsY          = canvas.Bounds().Max.Y - creditsFontHeight*2 - creditsFontSize/4
	)
	err = utils.DrawString(fc,
		creditsFont, memeFontSize, canvas.Bounds(),
		"",
		memeFontSpacing,
		[]string{creditsText},
		creditsX, creditsY)
	if err != nil {
		return nil, fmt.Errorf("DrawString (3): %w", err)
	}

	opts := jpeg.Options{Quality: 99}
	rw := bytes.Buffer{}
	err = jpeg.Encode(&rw, canvas, &opts)
	if err != nil {
		return nil, fmt.Errorf("jpeg.Encode: %w", err)
	}

	return &rw, nil
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
