package main

import (
	"fmt"
	"strings"

	"github.com/kyokomi/emoji/v2"
)

// resolveEmoji turns an argument into the literal emoji string. ':alias:'-style
// arguments are emojized via the gemoji alias set; anything else is treated as a
// raw UTF-8 emoji.
func resolveEmoji(arg string) (string, error) {
	var s string
	if len(arg) >= 2 && strings.HasPrefix(arg, ":") && strings.HasSuffix(arg, ":") {
		// kyokomi pads the resolved emoji with a trailing space so it renders
		// cleanly inline; trim it so it doesn't become a stray space codepoint.
		s = strings.TrimRight(emoji.Sprint(arg), " ")
		// kyokomi leaves unknown aliases untouched, so a still-colon-wrapped
		// result means the alias was not recognized.
		if s == arg {
			return "", fmt.Errorf("%s is not recognized as an emoji", arg)
		}
	} else {
		s = arg
	}
	if s == "" {
		return "", fmt.Errorf("%s is not recognized as an emoji", arg)
	}
	return s, nil
}

// codepointHex joins every rune's hex codepoint with '-'. An emoji can span
// several codepoints (e.g. a base char plus a U+FE0F variation selector), so we
// iterate runes rather than assuming a single codepoint.
func codepointHex(s string) string {
	parts := make([]string, 0, len(s))
	for _, r := range s {
		parts = append(parts, fmt.Sprintf("%x", r))
	}
	return strings.Join(parts, "-")
}
