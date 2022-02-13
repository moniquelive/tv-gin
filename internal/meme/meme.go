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
	_ "image/png" // enable reading of PNG files
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

func (mm memes) FindMeme(memeID string) (*meme, error) {
	for _, meme := range mm {
		if meme.ID == memeID {
			return &meme, nil
		}
	}
	return nil, fmt.Errorf("meme nao encontrado: %q", memeID)
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
	draw.Draw(canvas, canvas.Bounds(), img, img.Bounds().Min, draw.Src) // draw canvas background image (the meme)

	if len(texts) > 0 {
		rects := make([]image.Rectangle, 0, len(m.Boxes))
		for _, box := range m.Boxes {
			rects = append(rects, image.Rectangle{
				Min: image.Point{X: box[0], Y: box[1]},
				Max: image.Point{X: box[2], Y: box[3]},
			})
		}

		// DEBUG!
		// red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
		// for _, r := range m.Boxes {
		//	utils.Rect(canvas, red, r[0], r[1], r[2], r[3])
		//}

		const fontSpacing = 1
		for i, rect := range rects {
			wrappedText := wordWrap(texts[i], m.LineChars)
			memeFontSize := float64(utils.CalcMonoFontSize(memeFont, fontSpacing, wrappedText, rect))
			fc := utils.CreateFontContext(memeFont, memeFontSize, canvas.Bounds(), canvas, parseRGBA(m.FontColor))

			// Draw the text.
			y := utils.TextHeight(fc, memeFontSize, fontSpacing, wrappedText)
			err = utils.DrawString(fc, memeFont, memeFontSize, rect, m.TextAlign, fontSpacing, wrappedText,
				rect.Min.X+m.MarginLeft,
				rect.Min.Y+(rect.Dy()-y)/2)
			if err != nil {
				return nil, fmt.Errorf("DrawString: %w", err)
			}
		}

		if err := m.drawCredits(canvas); err != nil {
			return nil, fmt.Errorf("drawCredits: %w", err)
		}
	}

	rw := bytes.Buffer{}
	err = jpeg.Encode(&rw, canvas, &jpeg.Options{Quality: 99})
	if err != nil {
		return nil, fmt.Errorf("jpeg.Encode: %w", err)
	}

	return &rw, nil
}

func (m meme) drawCredits(canvas draw.Image) error {
	const (
		fontSpacing     = 1
		creditsFontSize = 32
		creditsText     = "Esta imagem foi gerada no meme.monique.dev" //nolint:gosec
	)
	var (
		creditsFontColor  = color.RGBA{R: 0xfe, G: 0x43, B: 0x65, A: 0xff}
		fc                = utils.CreateFontContext(creditsFont, creditsFontSize, canvas.Bounds(), canvas, creditsFontColor)
		creditsWidth      = utils.TextWidthInPixels(creditsFont, creditsFontSize, creditsText)
		creditsX          = canvas.Bounds().Max.X - creditsWidth - 12
		creditsFontHeight = int(fc.PointToFixed(creditsFontSize) >> 6)
		creditsY          = canvas.Bounds().Max.Y - int(float64(creditsFontHeight)*1.5)
	)
	err := utils.DrawString(fc,
		creditsFont, creditsFontSize, canvas.Bounds(),
		"",
		fontSpacing,
		[]string{creditsText},
		creditsX, creditsY)
	if err != nil {
		return fmt.Errorf("DrawString: %w", err)
	}
	return nil
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

func parseRGBA(rgba string) color.RGBA {
	c, err := utils.ParseHexColor(rgba)
	if err != nil {
		log.Panicf("parseRGBA> %v", err)
	}
	return c
}
