package meme

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"strings"

	"github.com/moniquelive/tv-gin/internal/utils"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type (
	meme struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		Filename   string   `json:"filename"`
		FontColor  string   `json:"font-color"`
		LineChars  int      `json:"line-chars"`
		MarginLeft int      `json:"margin-left"`
		TextAlign  string   `json:"text-align"`
		Boxes      [][4]int `json:"boxes"`
	}
	memes map[string]meme
)

var (
	//go:embed jetbrains.ttf
	fontBytes []byte
	memeFont  *truetype.Font

	//go:embed Handlee-Regular.ttf
	creditsFontBytes []byte
	creditsFont      *truetype.Font

	fsReader fs.ReadFileFS
)

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
}

func NewMeme(reader fs.ReadFileFS) *memes {
	fsReader = reader
	memesBytes, err := fsReader.ReadFile("web/config.json")
	if err != nil {
		log.Fatalf("NewMeme>ReadFile> %v", err)
	}
	var config memes
	if err := json.Unmarshal(memesBytes, &config); err != nil {
		log.Fatalln(err)
	}
	return &config
}

func (mm memes) FindMeme(name string) (*meme, error) {
	for _, meme := range mm {
		if meme.ID == name {
			return &meme, nil
		}
	}
	return nil, fmt.Errorf("meme nao encontrado: %q", name)
}

func (m meme) FontRGBA() color.RGBA {
	c, err := utils.ParseHexColor(m.FontColor)
	if err != nil {
		log.Panicf("FontRGBA> %v", err)
	}
	return c
}

func (m meme) Generate(texts []string) (*bytes.Buffer, error) {
	f, err := fsReader.Open("web/" + m.Filename)
	if err != nil {
		return nil, fmt.Errorf("fs.Open: %w", err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("image.Decode: %w", err)
	}
	canvas := image.NewRGBA(img.Bounds())
	draw.Draw(canvas, canvas.Bounds(), img, img.Bounds().Min, draw.Src)

	if len(texts) > 0 {
		coords := m.Boxes
		rects := make([]image.Rectangle, len(m.Boxes))
		for i := 0; i < len(m.Boxes); i++ {
			rects[i] = image.Rectangle{
				Min: image.Point{X: coords[i][0], Y: coords[i][1]},
				Max: image.Point{X: coords[i][2], Y: coords[i][3]},
			}
		}

		// DEBUG!
		//red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
		//for _, r := range m.Boxes {
		//	utils.Rect(canvas, red, r[0], r[1], r[2], r[3])
		//}

		const fontSpacing = 1
		for i := 0; i < len(m.Boxes); i++ {
			wrap := wordWrap(texts[i], m.LineChars)
			memeFontSize := float64(utils.CalcMonoFontSize(memeFont, fontSpacing, wrap, rects[i]))
			fc := utils.CreateFontContext(memeFont, memeFontSize, canvas.Bounds(), canvas, m.FontRGBA())

			// Draw the text.
			y := utils.TextHeight(fc, memeFontSize, fontSpacing, wrap)
			rectHalfHeight := rects[i].Dy() / 2
			err = utils.DrawString(fc,
				memeFont, memeFontSize, rects[i],
				m.TextAlign,
				fontSpacing,
				wrap,
				rects[i].Min.X+m.MarginLeft,
				rects[i].Min.Y+rectHalfHeight-y/2)
			if err != nil {
				return nil, fmt.Errorf("DrawString (1): %w", err)
			}
		}

		const (
			creditsFontSize = 32
			creditsText     = "Esta imagem foi gerada no meme.monique.dev"
		)
		creditsFontColor := color.RGBA{R: 0xfe, G: 0x43, B: 0x65, A: 0xff}
		fc := utils.CreateFontContext(creditsFont, creditsFontSize, canvas.Bounds(), canvas, creditsFontColor)
		var (
			creditsWidth      = utils.TextWidthInPixels(creditsFont, creditsFontSize, creditsText)
			creditsX          = canvas.Bounds().Max.X - creditsWidth - 12
			creditsFontHeight = int(fc.PointToFixed(creditsFontSize) >> 6)
			creditsY          = canvas.Bounds().Max.Y - int(float64(creditsFontHeight)*1.5)
		)
		err = utils.DrawString(fc,
			creditsFont, creditsFontSize, canvas.Bounds(),
			"",
			fontSpacing,
			[]string{creditsText},
			creditsX, creditsY)
		if err != nil {
			return nil, fmt.Errorf("DrawString (3): %w", err)
		}
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
