package utils

import (
	"fmt"
	"image/color"
	"log"
)

func ParseHexColor(s string) (c color.RGBA, err error) {
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length: must be 7 or 4")
	}
	return
}

func ParseRGBA(rgba string) color.RGBA {
	c, err := ParseHexColor(rgba)
	if err != nil {
		log.Panicf("ParseRGBA> %v", err)
	}
	return c
}
