package utils

import (
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var fontSizes [255]int

func init() {
	for i := 0; i < len(fontSizes); i++ {
		fontSizes[i] = i + 1
	}
}

func CreateFontContext(ttFont *truetype.Font, fontSize float64, clipRectangle image.Rectangle, canvas draw.Image, color color.Color) *freetype.Context {
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(ttFont)
	fc.SetFontSize(fontSize)
	fc.SetClip(clipRectangle)
	fc.SetDst(canvas)
	fc.SetSrc(image.NewUniform(color))
	fc.SetHinting(font.HintingNone)
	// fc.SetHinting(font.HintingFull)
	return fc
}

// DrawString calculates the widths and print to image
func DrawString(fc *freetype.Context, ttFont *truetype.Font, fontSize float64, clipRectangle image.Rectangle, alignment string, spacing float64, text []string, x, y int) error {
	fontHeight := int(fc.PointToFixed(fontSize) >> 6)
	pt := freetype.Pt(x, y+fontHeight)
	for _, s := range text {
		width := TextWidthInPixels(ttFont, fontSize, s)
		_, err := fc.DrawString(s, align(alignment, width, clipRectangle, pt))
		if err != nil {
			log.Println(err)
			return err
		}
		pt.Y += fc.PointToFixed(fontSize * spacing)
	}
	return nil
}

func align(alignment string, width int, clipRectangle image.Rectangle, pt fixed.Point26_6) fixed.Point26_6 {
	x, y := pt.X, pt.Y
	if strings.ToLower(alignment) == "center" {
		rectPt := freetype.Pt((clipRectangle.Max.X-clipRectangle.Min.X-width)/2, 0)
		x += rectPt.X
	}
	return fixed.Point26_6{X: x, Y: y}
}

func TextHeight(fc *freetype.Context, size, spacing float64, lines []string) int {
	return int(fc.PointToFixed(size)>>6) +
		(len(lines)-1)*(int(fc.PointToFixed(size*spacing)>>6))
}

func TextWidthInPixels(ttFont *truetype.Font, size float64, text string) int {
	face := truetype.NewFace(ttFont, &truetype.Options{Size: size})
	return font.MeasureString(face, text).Floor()
}

// CalcMonoFontSize calcula o tamanho maximo da fonte usada no bloco de texto para caber em bounds.
func CalcMonoFontSize(ttFont *truetype.Font, spacing float64, textLines []string, bounds image.Rectangle) int {
	return recCalcMonoFontSize(freetype.NewContext(), fontSizes[:], ttFont, spacing, textLines, bounds)
}

func recCalcMonoFontSize(fc *freetype.Context, sizes []int, ttFont *truetype.Font, spacing float64, textLines []string, bounds image.Rectangle) int {
	if len(sizes) == 1 {
		return sizes[0]
	}
	// fmt.Println(sizes)
	currSize := sizes[len(sizes)/2]
	_, advWidth, _ := truetype.NewFace(ttFont, &truetype.Options{Size: float64(currSize)}).GlyphBounds(' ')
	charWidthInPixels := advWidth.Round()

	height := TextHeight(fc, float64(currSize), spacing, textLines)
	var maxWidth float64
	for _, w := range textLines {
		maxWidth = math.Max(maxWidth, float64(len(w)*charWidthInPixels))
	}
	if height > bounds.Dy() || maxWidth > float64(bounds.Dx()) { // too large for bounds, shrink
		return recCalcMonoFontSize(fc, sizes[:len(sizes)/2], ttFont, spacing, textLines, bounds)
	}
	// too small for bounds, grow
	return recCalcMonoFontSize(fc, sizes[len(sizes)/2:], ttFont, spacing, textLines, bounds)
}

func WordWrap(text string, lineWidth int) (lines []string) {
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
