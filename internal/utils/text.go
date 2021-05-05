package utils

import (
	"image"
	"image/color"
	"log"
	"math"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

var allSizes []int

func init() {
	allSizes = make([]int, 256)
	for i := 0; i < 256; i++ {
		allSizes[i] = i + 1
	}

}

func CreateFontContext(ttFont *truetype.Font, fontSize float64, clipRectangle image.Rectangle, canvas *image.RGBA, color color.RGBA) *freetype.Context {
	fc := freetype.NewContext()
	fc.SetDPI(72)
	fc.SetFont(ttFont)
	fc.SetFontSize(fontSize)
	fc.SetClip(clipRectangle)
	fc.SetDst(canvas)
	fc.SetSrc(image.NewUniform(color))
	fc.SetHinting(font.HintingNone)
	//fc.SetHinting(font.HintingFull)
	return fc
}

func DrawString(
	fc *freetype.Context,
	ttFont *truetype.Font,
	fontSize float64,
	clipRectangle image.Rectangle,
	alignment string,
	spacing float64,
	text []string,
	x, y int) error {

	// Calculate the widths and print to image
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
	x := pt.X
	y := pt.Y
	switch strings.ToLower(alignment) {
	case "center":
		rectPt := freetype.Pt((clipRectangle.Max.X-clipRectangle.Min.X-width)/2, 0)
		x += rectPt.X
	}
	return fixed.Point26_6{X: x, Y: y}
}

func TextHeight(fc *freetype.Context, size, spacing float64, lines []string) int {
	return int(fc.PointToFixed(size)>>6) +
		(len(lines)-1)*(int(fc.PointToFixed(size*spacing)>>6))
}

func TextWidthInPixels(f *truetype.Font, size float64, text string) int {
	opts := truetype.Options{Size: size}
	face := truetype.NewFace(f, &opts)
	return font.MeasureString(face, text).Floor()
}

// CalcMonoFontSize calcula o tamanho maximo da fonte usada no bloco de texto para caber em bounds.
func CalcMonoFontSize(f *truetype.Font, spacing float64, wrap []string, bounds image.Rectangle) int {
	return recCalcMonoFontSize(freetype.NewContext(), allSizes, f, spacing, wrap, bounds)
}

func recCalcMonoFontSize(fc *freetype.Context, sizes []int, f *truetype.Font, spacing float64, wrap []string, bounds image.Rectangle) int {
	//fmt.Println(sizes)
	currSize := sizes[len(sizes)/2]
	opts := truetype.Options{Size: float64(currSize)}
	face := truetype.NewFace(f, &opts)
	_, a, _ := face.GlyphBounds(' ')
	charWidthInPixels := a.Round()

	height := TextHeight(fc, float64(currSize), spacing, wrap)
	maxWidth := float64(0)
	for _, w := range wrap {
		maxWidth = math.Max(maxWidth, float64(len(w)*charWidthInPixels))
	}
	if len(sizes) == 1 {
		return sizes[0]
	}
	if height > bounds.Dy() || maxWidth > float64(bounds.Dx()) {
		firstHalf := sizes[0 : len(sizes)/2]
		return recCalcMonoFontSize(fc, firstHalf, f, spacing, wrap, bounds)
	} else {
		secondHalf := sizes[len(sizes)/2:]
		return recCalcMonoFontSize(fc, secondHalf, f, spacing, wrap, bounds)
	}
}
