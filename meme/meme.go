package meme

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png" // enable reading of PNG files
	"log"

	"github.com/moniquelive/tv-gin/utils"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

type (
	Meme struct {
		ID         string   `json:"id"`
		Name       string   `json:"name"`
		Filename   string   `json:"filename"`
		FontColor  string   `json:"font-color"`
		LineChars  int      `json:"line-chars"`
		MarginLeft int      `json:"margin-left"`
		TextAlign  string   `json:"text-align"`
		Boxes      [][4]int `json:"boxes"`
	}
	Memes map[string]Meme
)

var (
	//go:embed MonaspaceRadonVarVF.ttf
	fontBytes []byte
	memeFont  *truetype.Font

	//go:embed Handlee-Regular.ttf
	creditsFontBytes []byte
	creditsFont      *truetype.Font
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

func NewMeme(configFile []byte) (config *Memes, err error) {
	err = json.Unmarshal(configFile, &config)
	return
}

func (mm Memes) FindMeme(memeID string) (*Meme, error) {
	for _, m := range mm {
		if m.ID == memeID {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("meme nao encontrado: %q", memeID)
}

func (m Meme) Generate(webRoot embed.FS, texts []string) (*bytes.Buffer, error) {
	f, err := webRoot.Open("web/" + m.Filename)
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
		if err := m.DrawTexts(texts, canvas); err != nil {
			return nil, err
		}
	}

	if err := m.drawCredits(canvas); err != nil {
		return nil, fmt.Errorf("drawCredits: %w", err)
	}

	rw := bytes.Buffer{}
	err = jpeg.Encode(&rw, canvas, &jpeg.Options{Quality: 99})
	if err != nil {
		return nil, fmt.Errorf("jpeg.Encode: %w", err)
	}

	return &rw, nil
}

func (m Meme) DrawTexts(texts []string, canvas *image.RGBA) (err error) {
	rects := make([]image.Rectangle, 0, len(m.Boxes))
	for _, box := range m.Boxes {
		rectangle := image.Rectangle{
			Min: image.Point{X: box[0], Y: box[1]},
			Max: image.Point{X: box[2], Y: box[3]},
		}
		rects = append(rects, rectangle)
	}

	// DEBUG!
	// red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	// for _, r := range m.Boxes {
	//	utils.Rect(canvas, red, r[0], r[1], r[2], r[3])
	//}

	const fontSpacing = 1
	for i, rect := range rects {
		wrappedText := utils.WordWrap(texts[i], m.LineChars)
		memeFontSize := float64(utils.CalcMonoFontSize(memeFont, fontSpacing, wrappedText, rect))
		fc := utils.CreateFontContext(memeFont, memeFontSize, canvas.Bounds(), canvas, utils.ParseRGBA(m.FontColor))

		// Draw the text.
		y := utils.TextHeight(fc, memeFontSize, fontSpacing, wrappedText)
		err = utils.DrawString(fc, memeFont, memeFontSize, rect, m.TextAlign, fontSpacing, wrappedText,
			rect.Min.X+m.MarginLeft,
			rect.Min.Y+(rect.Dy()-y)/2)
		if err != nil {
			return fmt.Errorf("DrawString: %w", err)
		}
	}
	return
}

func (m Meme) drawCredits(canvas draw.Image) (err error) {
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
	err = utils.DrawString(fc,
		creditsFont, creditsFontSize, canvas.Bounds(),
		"",
		fontSpacing,
		[]string{creditsText},
		creditsX, creditsY)
	if err != nil {
		return fmt.Errorf("DrawString: %w", err)
	}
	return
}
