package meme

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"log"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed static/jetbrains.ttf
var fontBytes []byte

//go:embed static/Handlee-Regular.ttf
var creditsFontBytes []byte

var memeFont *truetype.Font
var creditsFont *truetype.Font

type Meme interface {
	Generate([]string) (*bytes.Buffer, error)
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
}

func New(name string) (Meme, error) {
	switch name {
	case "drake":
		return &drake{}, nil
	default:
		return nil, fmt.Errorf("meme desconhecido: %q", name)
	}
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

func measureString(fc *freetype.Context, size, spacing float64, text []string) int {
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
