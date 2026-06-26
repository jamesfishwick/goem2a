package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// ansiPattern matches a single SGR escape sequence, e.g.
// "\x1b[38;2;96;96;96m" or "\x1b[0m".
var ansiPattern = regexp.MustCompile("\x1b\\[([0-9;]*)m")

// updateColor applies an SGR parameter list to the current color. jp2a 1.3.3
// only ever emits truecolor foregrounds ("38;2;r;g;b") and resets ("0"), so we
// parse just those and treat anything else as a no-op.
func updateColor(current color.RGBA, codes string) color.RGBA {
	parts := splitInts(codes)
	if len(parts) == 0 {
		parts = []int{0}
	}
	for i := 0; i < len(parts); {
		switch {
		case parts[i] == 0:
			current = color.RGBA{0, 0, 0, 255}
			i++
		case parts[i] == 38 && i+4 < len(parts) && parts[i+1] == 2:
			current = color.RGBA{uint8(parts[i+2]), uint8(parts[i+3]), uint8(parts[i+4]), 255}
			i += 5
		default:
			i++
		}
	}
	return current
}

func splitInts(s string) []int {
	var out []int
	for _, p := range strings.Split(s, ";") {
		if p == "" {
			continue
		}
		if n, err := strconv.Atoi(p); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// renderASCIIToPNG rasterizes jp2a's (possibly ANSI-colored) ASCII art to a PNG
// using an embedded monospace font, so every glyph occupies one fixed cell.
func renderASCIIToPNG(asciiArt, outPath string, bg background) error {
	const fontSize = 14
	face, err := newMonoFace(fontSize)
	if err != nil {
		return err
	}
	defer face.Close()

	metrics := face.Metrics()
	ascent := metrics.Ascent.Ceil()
	cellHeight := (metrics.Ascent + metrics.Descent).Ceil()
	cellWidth := font.MeasureString(face, "M").Ceil()
	if cellWidth == 0 {
		cellWidth = fontSize
	}

	lines := strings.Split(asciiArt, "\n")
	cols := 1
	for _, line := range lines {
		if n := len([]rune(ansiPattern.ReplaceAllString(line, ""))); n > cols {
			cols = n
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, cols*cellWidth, len(lines)*cellHeight))
	// 'transparent' leaves the RGBA canvas fully clear; any other value is an
	// opaque fill.
	if !bg.transparent {
		draw.Draw(img, img.Bounds(), &image.Uniform{bg.fill}, image.Point{}, draw.Src)
	}

	drawer := &font.Drawer{Dst: img, Face: face}
	// paintRun draws each rune of s into successive cells starting at col,
	// returning the next free column.
	paintRun := func(s string, col int, baseY fixed.Int26_6, c color.RGBA) int {
		for _, ch := range s {
			drawer.Src = &image.Uniform{c}
			drawer.Dot = fixed.Point26_6{X: fixed.I(col * cellWidth), Y: baseY}
			drawer.DrawString(string(ch))
			col++
		}
		return col
	}

	for row, line := range lines {
		col := 0
		curColor := color.RGBA{0, 0, 0, 255}
		baseY := fixed.I(row*cellHeight + ascent)
		pos := 0
		for _, loc := range ansiPattern.FindAllStringSubmatchIndex(line, -1) {
			start, end, gStart, gEnd := loc[0], loc[1], loc[2], loc[3]
			col = paintRun(line[pos:start], col, baseY, curColor)
			curColor = updateColor(curColor, line[gStart:gEnd])
			pos = end
		}
		paintRun(line[pos:], col, baseY, curColor)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	// Check Close explicitly: png.Encode can succeed while the final flush on
	// Close fails, which would otherwise report a truncated PNG as success.
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func newMonoFace(size float64) (font.Face, error) {
	parsed, err := opentype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}
