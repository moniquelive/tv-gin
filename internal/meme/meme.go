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
	"log"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
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
	Name       string   `json:"name"`
	Filename   string   `json:"filename"`
	MarginLeft int      `json:"margin-left"`
	Boxes      [][4]int `json:"boxes"`
}

var config struct {
	Memes map[string]meme `json:"memes"`
}

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
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		log.Fatalln(err)
	}
}

func Generate(name string, texts []string) (*bytes.Buffer, error) {
	var meme meme
	meme, ok := config.Memes[name]
	if !ok {
		return nil, fmt.Errorf("meme não encontrado: %q", name)
	}
	coords := meme.Boxes
	rects := [2]image.Rectangle{
		{image.Point{X: coords[0][0], Y: coords[0][1]}, image.Point{X: coords[0][2], Y: coords[0][3]}},
		{image.Point{X: coords[1][0], Y: coords[1][1]}, image.Point{X: coords[1][2], Y: coords[1][3]}},
	}
	f, err := os.Open("./testdata/" + meme.Filename)
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

	const (
		memeFontSize    = 64
		memeFontSpacing = 1
		wordwrapWidth   = 13
	)
	fc := createFontContext(memeFont, memeFontSize, canvas.Bounds(), canvas, image.Black)

	// Draw the text.
	y := textHeight(fc, memeFontSize, memeFontSpacing, wordWrap(texts[0], wordwrapWidth))
	rect0HalfHeight := rects[0].Dy() / 2
	err = drawString(fc,
		memeFontSize, memeFontSpacing,
		wordWrap(texts[0], wordwrapWidth),
		rects[0].Min.X+meme.MarginLeft,
		rects[0].Min.Y+rect0HalfHeight-y/2)
	if err != nil {
		return nil, fmt.Errorf("drawString (1): %w", err)
	}

	y = textHeight(fc, memeFontSize, memeFontSpacing, wordWrap(texts[1], wordwrapWidth))
	rect1HalfHeight := rects[1].Dy() / 2
	err = drawString(fc,
		memeFontSize, memeFontSpacing,
		wordWrap(texts[1], wordwrapWidth),
		rects[1].Min.X+meme.MarginLeft,
		rects[1].Min.Y+rect1HalfHeight-y/2)
	if err != nil {
		return nil, fmt.Errorf("drawString (2): %w", err)
	}

	const (
		creditsFontSize = 32
		creditsText     = "Esta imagem foi gerada no meme.monique.dev"
	)
	creditsFontColor := image.NewUniform(color.RGBA{R: 0xfe, G: 0x43, B: 0x65, A: 0xff})
	fc = createFontContext(creditsFont, creditsFontSize, canvas.Bounds(), canvas, creditsFontColor)
	var (
		creditsWidth      = creditsWidthInPixels(creditsFont, creditsFontSize, creditsText)
		creditsX          = canvas.Bounds().Max.X - creditsWidth - creditsFontSize
		creditsFontHeight = int(fc.PointToFixed(creditsFontSize) >> 6)
		creditsY          = canvas.Bounds().Max.Y - creditsFontHeight*2 - creditsFontSize/4
	)
	err = drawString(fc, memeFontSize, memeFontSpacing, []string{creditsText}, creditsX, creditsY)
	if err != nil {
		return nil, fmt.Errorf("drawString (3): %w", err)
	}

	opts := jpeg.Options{Quality: 99}
	rw := bytes.Buffer{}
	err = jpeg.Encode(&rw, canvas, &opts)
	if err != nil {
		return nil, fmt.Errorf("jpeg.Encode: %w", err)
	}

	return &rw, nil
}

func createFontContext(ttFont *truetype.Font, fontSize float64, clipRectangle image.Rectangle, canvas *image.RGBA, color *image.Uniform) *freetype.Context {
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(ttFont)
	fc.SetFontSize(fontSize)
	fc.SetClip(clipRectangle)
	fc.SetDst(canvas)
	fc.SetSrc(color)
	fc.SetHinting(font.HintingNone)
	//fc.SetHinting(font.HintingFull)
	return fc
}

func drawString(fc *freetype.Context, size, spacing float64, text []string, x, y int) error {
	// Calculate the widths and print to image
	fontHeight := int(fc.PointToFixed(size) >> 6)
	pt := freetype.Pt(x, y+fontHeight)
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

func creditsWidthInPixels(f *truetype.Font, size float64, text string) int {
	opts := truetype.Options{
		Size: size,
	}
	face := truetype.NewFace(f, &opts)
	width := 0
	for _, x := range text {
		awidth, _ := face.GlyphAdvance(x)
		iwidthf := int(float64(awidth) / 64)
		width += iwidthf
	}
	return width
}

func textHeight(fc *freetype.Context, size, spacing float64, text []string) int {
	return int(fc.PointToFixed(size)>>6) +
		(len(text)-1)*(int(fc.PointToFixed(size*spacing)>>6))
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
