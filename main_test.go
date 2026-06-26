package main

import (
	"image/color"
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantColor   bool
		wantMirror  bool
		wantPNG     bool
		wantBG      string // expected jp2aBackground; "" means don't check
		wantRest    []string
		wantErr     bool
	}{
		{
			name:      "plain emoji",
			args:      []string{"\U0001F525"},
			wantRest:  []string{"\U0001F525"},
		},
		{
			name:      "color flag",
			args:      []string{"--color", ":smile:"},
			wantColor: true,
			wantRest:  []string{":smile:"},
		},
		{
			name:      "mirror flag",
			args:      []string{"--mirror", "x"},
			wantMirror: true,
			wantRest:  []string{"x"},
		},
		{
			name:      "png implies color",
			args:      []string{"--png", "x"},
			wantColor: true,
			wantPNG:   true,
			wantRest:  []string{"x"},
		},
		{
			name:      "bg space-separated implies png and color, dark ramp",
			args:      []string{"--bg", "black", "x"},
			wantColor: true,
			wantPNG:   true,
			wantBG:    "dark",
			wantRest:  []string{"x"},
		},
		{
			name:      "bg equals-form, light ramp",
			args:      []string{"--bg=white", "x"},
			wantColor: true,
			wantPNG:   true,
			wantBG:    "light",
			wantRest:  []string{"x"},
		},
		{
			name:     "multiple emojis preserved in order",
			args:     []string{"a", "--color", "b"},
			wantColor: true,
			wantRest: []string{"a", "b"},
		},
		{
			name:    "bg without value errors",
			args:    []string{"--bg"},
			wantErr: true,
		},
		{
			name:    "bg with unknown color errors",
			args:    []string{"--bg", "purplish", "x"},
			wantErr: true,
		},
		{
			name:    "no emoji arguments errors",
			args:    []string{"--color"},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts, rest, err := parseArgs(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseArgs(%v) = no error, want error", tc.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseArgs(%v) unexpected error: %v", tc.args, err)
			}
			if opts.color != tc.wantColor {
				t.Errorf("color = %v, want %v", opts.color, tc.wantColor)
			}
			if opts.mirror != tc.wantMirror {
				t.Errorf("mirror = %v, want %v", opts.mirror, tc.wantMirror)
			}
			if opts.png != tc.wantPNG {
				t.Errorf("png = %v, want %v", opts.png, tc.wantPNG)
			}
			if tc.wantBG != "" && opts.jp2aBackground != tc.wantBG {
				t.Errorf("jp2aBackground = %q, want %q", opts.jp2aBackground, tc.wantBG)
			}
			if !reflect.DeepEqual(rest, tc.wantRest) {
				t.Errorf("rest = %v, want %v", rest, tc.wantRest)
			}
		})
	}
}

func TestParseArgsDefaults(t *testing.T) {
	opts, _, err := parseArgs([]string{"x"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.jp2aBackground != "light" {
		t.Errorf("default jp2aBackground = %q, want light", opts.jp2aBackground)
	}
	if opts.bg.transparent {
		t.Error("default background should not be transparent")
	}
	if opts.bg.fill != (color.RGBA{255, 255, 255, 255}) {
		t.Errorf("default fill = %v, want white", opts.bg.fill)
	}
}

func TestParseBackground(t *testing.T) {
	tests := []struct {
		name            string
		value           string
		wantTransparent bool
		wantRamp        string
		wantErr         bool
	}{
		{"transparent keeps light ramp", "transparent", true, "light", false},
		{"none alias", "none", true, "light", false},
		{"dark color picks dark ramp", "black", false, "dark", false},
		{"light color picks light ramp", "white", false, "light", false},
		{"dark hex picks dark ramp", "#1e1e2e", false, "dark", false},
		{"unknown errors", "bogus", false, "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := options{jp2aBackground: "light"}
			err := parseBackground(tc.value, &opts)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseBackground(%q) = no error, want error", tc.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseBackground(%q) unexpected error: %v", tc.value, err)
			}
			if opts.bg.transparent != tc.wantTransparent {
				t.Errorf("transparent = %v, want %v", opts.bg.transparent, tc.wantTransparent)
			}
			if opts.jp2aBackground != tc.wantRamp {
				t.Errorf("jp2aBackground = %q, want %q", opts.jp2aBackground, tc.wantRamp)
			}
		})
	}
}
