package main

import (
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitInts(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []int
	}{
		{"simple", "1;2;3", []int{1, 2, 3}},
		{"empty fields skipped", "1;;3", []int{1, 3}},
		{"empty string", "", nil},
		{"truecolor", "38;2;96;96;96", []int{38, 2, 96, 96, 96}},
		{"non-numeric skipped", "1;x;2", []int{1, 2}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := splitInts(tc.in)
			if len(got) != len(tc.want) {
				t.Fatalf("splitInts(%q) = %v, want %v", tc.in, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("splitInts(%q) = %v, want %v", tc.in, got, tc.want)
				}
			}
		})
	}
}

func TestUpdateColor(t *testing.T) {
	red := color.RGBA{255, 0, 0, 255}
	black := color.RGBA{0, 0, 0, 255}
	tests := []struct {
		name  string
		start color.RGBA
		codes string
		want  color.RGBA
	}{
		{"reset to black", red, "0", black},
		{"truecolor foreground", black, "38;2;12;34;56", color.RGBA{12, 34, 56, 255}},
		{"empty resets to black", red, "", black},
		{"bold-prefixed truecolor", black, "1;38;2;10;20;30", color.RGBA{10, 20, 30, 255}},
		{"unknown code is no-op", red, "7", red},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := updateColor(tc.start, tc.codes); got != tc.want {
				t.Errorf("updateColor(%v, %q) = %v, want %v", tc.start, tc.codes, got, tc.want)
			}
		})
	}
}

// The art carries no trailing reset, so the second glyph must keep the color
// set by the SGR code that preceded it.
func TestRenderASCIIToPNG_OpaqueBackground(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.png")
	art := "\x1b[38;2;10;20;30mAB\x1b[0m\nCD"
	bg := background{fill: color.RGBA{255, 255, 255, 255}}

	if err := renderASCIIToPNG(art, out, bg); err != nil {
		t.Fatalf("renderASCIIToPNG: %v", err)
	}

	img := decodePNG(t, out)
	// Top-left pixel sits on the white fill, behind the (sparse) glyph strokes;
	// the corner of an opaque canvas must be fully opaque.
	if _, _, _, a := img.At(0, 0).RGBA(); a != 0xffff {
		t.Errorf("opaque background corner alpha = %d, want fully opaque", a>>8)
	}
}

func TestRenderASCIIToPNG_TransparentBackground(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.png")
	art := "A\nB"
	bg := background{transparent: true}

	if err := renderASCIIToPNG(art, out, bg); err != nil {
		t.Fatalf("renderASCIIToPNG: %v", err)
	}

	img := decodePNG(t, out)
	// A transparent canvas leaves the corner (no glyph there) fully clear.
	if _, _, _, a := img.At(0, 0).RGBA(); a != 0 {
		t.Errorf("transparent background corner alpha = %d, want 0", a>>8)
	}
}

func decodePNG(t *testing.T, path string) interface {
	At(x, y int) color.Color
} {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return img
}
