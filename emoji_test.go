package main

import "testing"

func TestResolveEmoji(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"known alias", ":smile:", "\U0001F604", false},
		{"raw single emoji", "\U0001F525", "\U0001F525", false},
		{"raw multi-codepoint emoji", "\U0001F5C3️", "\U0001F5C3️", false},
		{"unknown alias", ":notanemoji:", "", true},
		{"empty string", "", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveEmoji(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("resolveEmoji(%q) = %q, want error", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveEmoji(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Errorf("resolveEmoji(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// Aliases must not leak kyokomi's trailing padding space into the codepoint
// sequence (which would become a stray "-20" and break the CDN lookup).
func TestResolveEmojiHasNoTrailingSpace(t *testing.T) {
	got, err := resolveEmoji(":smile:")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[len(got)-1] == ' ' {
		t.Errorf("resolveEmoji(\":smile:\") = %q has a trailing space", got)
	}
}

func TestCodepointHex(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"ascii letter", "A", "41"},
		{"single emoji", "\U0001F525", "1f525"},
		{"multi-codepoint with variation selector", "\U0001F5C3️", "1f5c3-fe0f"},
		{"empty", "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := codepointHex(tc.in); got != tc.want {
				t.Errorf("codepointHex(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
