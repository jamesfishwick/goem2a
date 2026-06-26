package main

import (
	"image/color"
	"math"
	"testing"
)

func TestParseColor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    color.RGBA
		wantErr bool
	}{
		{"named", "white", color.RGBA{255, 255, 255, 255}, false},
		{"named case-insensitive", "BLACK", color.RGBA{0, 0, 0, 255}, false},
		{"named with surrounding space", "  red  ", color.RGBA{255, 0, 0, 255}, false},
		{"grey alias", "grey", color.RGBA{128, 128, 128, 255}, false},
		{"hex 6-digit", "#1e1e2e", color.RGBA{30, 30, 46, 255}, false},
		{"hex shorthand", "#abc", color.RGBA{170, 187, 204, 255}, false},
		{"unknown name", "purplish", color.RGBA{}, true},
		{"bad hex length", "#12345", color.RGBA{}, true},
		{"bad hex digits", "#zzzzzz", color.RGBA{}, true},
		{"empty", "", color.RGBA{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseColor(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseColor(%q) = %v, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseColor(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("parseColor(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestLuminance(t *testing.T) {
	tests := []struct {
		name string
		in   color.RGBA
		want float64
	}{
		{"white is bright", color.RGBA{255, 255, 255, 255}, 1.0},
		{"black is dark", color.RGBA{0, 0, 0, 255}, 0.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Rec.601 coefficients don't sum to exactly 1.0 in float64, so
			// compare with a small tolerance rather than for exact equality.
			if got := luminance(tc.in); math.Abs(got-tc.want) > 1e-9 {
				t.Errorf("luminance(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}

	// Green carries more luminance weight than blue at equal intensity, so a
	// pure-green canvas reads as "light" and a pure-blue one as "dark".
	if luminance(color.RGBA{0, 255, 0, 255}) < 0.5 {
		t.Error("pure green should be treated as a light canvas")
	}
	if luminance(color.RGBA{0, 0, 255, 255}) >= 0.5 {
		t.Error("pure blue should be treated as a dark canvas")
	}
}
