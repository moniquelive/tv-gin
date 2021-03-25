package meme

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
)

//go:embed static/drake.jpg
var memeBytes []byte

var (
	coords = [2][4]int{
		{600, 0, 1199, 599},
		{600, 600, 1199, 1199},
	}
	leftMargin = 55
)

type drake struct {
}

func (drake) Generate(texts []string) (*bytes.Buffer, error) {
	rects := [2]image.Rectangle{
		{image.Point{X: coords[0][0], Y: coords[0][1]}, image.Point{X: coords[0][2], Y: coords[0][3]}},
		{image.Point{X: coords[1][0], Y: coords[1][1]}, image.Point{X: coords[1][2], Y: coords[1][3]}},
	}
	img, _, err := image.Decode(bytes.NewReader(memeBytes))
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
		rects[0].Min.X+leftMargin,
		rects[0].Min.Y+rect0HalfHeight-y/2)
	if err != nil {
		return nil, fmt.Errorf("drawString (1): %w", err)
	}

	y = textHeight(fc, memeFontSize, memeFontSpacing, wordWrap(texts[1], wordwrapWidth))
	rect1HalfHeight := rects[1].Dy() / 2
	err = drawString(fc,
		memeFontSize, memeFontSpacing,
		wordWrap(texts[1], wordwrapWidth),
		rects[1].Min.X+leftMargin,
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
