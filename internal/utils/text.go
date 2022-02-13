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

var fontSizes = []int{
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31,
	32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60,
	61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89,
	90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114,
	115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137,
	138, 139, 140, 141, 142, 143, 144, 145, 146, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159, 160,
	161, 162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180, 181, 182, 183,
	184, 185, 186, 187, 188, 189, 190, 191, 192, 193, 194, 195, 196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206,
	207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224, 225, 226, 227, 228, 229,
	230, 231, 232, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252,
	253, 254, 255}

func CreateFontContext(
	ttFont *truetype.Font,
	fontSize float64,
	clipRectangle image.Rectangle,
	canvas draw.Image,
	color color.Color) *freetype.Context {
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
func DrawString(
	fc *freetype.Context,
	ttFont *truetype.Font,
	fontSize float64,
	clipRectangle image.Rectangle,
	alignment string,
	spacing float64,
	text []string,
	x, y int) error {
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

func TextWidthInPixels(f *truetype.Font, size float64, text string) int {
	face := truetype.NewFace(f, &truetype.Options{Size: size})
	return font.MeasureString(face, text).Floor()
}

// CalcMonoFontSize calcula o tamanho maximo da fonte usada no bloco de texto para caber em bounds.
func CalcMonoFontSize(f *truetype.Font, spacing float64, textLines []string, bounds image.Rectangle) int {
	return recCalcMonoFontSize(
		freetype.NewContext(),
		fontSizes,
		f,
		spacing,
		textLines,
		bounds)
}

func recCalcMonoFontSize(
	fc *freetype.Context,
	sizes []int,
	ttf *truetype.Font,
	spacing float64,
	textLines []string,
	bounds image.Rectangle) int {
	// fmt.Println(sizes)
	currSize := sizes[len(sizes)/2]
	_, advWidth, _ := truetype.NewFace(ttf, &truetype.Options{Size: float64(currSize)}).GlyphBounds(' ')
	charWidthInPixels := advWidth.Round()

	height := TextHeight(fc, float64(currSize), spacing, textLines)
	var maxWidth float64
	for _, w := range textLines {
		maxWidth = math.Max(maxWidth, float64(len(w)*charWidthInPixels))
	}
	if len(sizes) == 1 {
		return sizes[0]
	}
	if height > bounds.Dy() || maxWidth > float64(bounds.Dx()) { // too large for bounds, shrink
		return recCalcMonoFontSize(fc, sizes[:len(sizes)/2], ttf, spacing, textLines, bounds)
	}
	// too small for bounds, grow
	return recCalcMonoFontSize(fc, sizes[len(sizes)/2:], ttf, spacing, textLines, bounds)
}
