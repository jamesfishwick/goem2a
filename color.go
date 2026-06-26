package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

// cssColors is a practical subset of the CSS3 named colors. Pillow's
// ImageColor.getrgb accepts the full ~140-name CSS list; we cover the common
// ones and otherwise rely on #rrggbb, which is unambiguous.
var cssColors = map[string]color.RGBA{
	"black":   {0, 0, 0, 255},
	"white":   {255, 255, 255, 255},
	"red":     {255, 0, 0, 255},
	"green":   {0, 128, 0, 255},
	"lime":    {0, 255, 0, 255},
	"blue":    {0, 0, 255, 255},
	"yellow":  {255, 255, 0, 255},
	"cyan":    {0, 255, 255, 255},
	"magenta": {255, 0, 255, 255},
	"gray":    {128, 128, 128, 255},
	"grey":    {128, 128, 128, 255},
	"silver":  {192, 192, 192, 255},
	"maroon":  {128, 0, 0, 255},
	"olive":   {128, 128, 0, 255},
	"purple":  {128, 0, 128, 255},
	"teal":    {0, 128, 128, 255},
	"navy":    {0, 0, 128, 255},
	"orange":  {255, 165, 0, 255},
	"pink":    {255, 192, 203, 255},
	"brown":   {165, 42, 42, 255},
	"gold":    {255, 215, 0, 255},
	"indigo":  {75, 0, 130, 255},
	"violet":  {238, 130, 238, 255},
}

// parseColor resolves a color name or #rrggbb / #rgb string to an RGBA value.
func parseColor(value string) (color.RGBA, error) {
	v := strings.TrimSpace(value)
	if c, ok := cssColors[strings.ToLower(v)]; ok {
		return c, nil
	}
	if strings.HasPrefix(v, "#") {
		hex := v[1:]
		// #rgb shorthand expands each nibble (e.g. #1e2 -> #11ee22).
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) == 6 {
			n, err := strconv.ParseUint(hex, 16, 32)
			if err == nil {
				return color.RGBA{
					R: uint8(n >> 16),
					G: uint8(n >> 8),
					B: uint8(n),
					A: 255,
				}, nil
			}
		}
	}
	return color.RGBA{}, fmt.Errorf("%q is not a recognized color (try a name, #rrggbb, or \"transparent\")", value)
}

// luminance returns Rec. 601 relative luminance in [0,1]; a dark canvas wants
// jp2a's dark character ramp.
func luminance(c color.RGBA) float64 {
	return (0.299*float64(c.R) + 0.587*float64(c.G) + 0.114*float64(c.B)) / 255
}
