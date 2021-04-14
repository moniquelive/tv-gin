package utils

import (
	"image"
	"image/color"
	"log"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)

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

func TextHeight(fc *freetype.Context, size, spacing float64, text []string) int {
	return int(fc.PointToFixed(size)>>6) +
		(len(text)-1)*(int(fc.PointToFixed(size*spacing)>>6))
}

func TextWidthInPixels(f *truetype.Font, size float64, text string) int {
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
