package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// emojiCDN is the base URL for Apple-style emoji PNGs. It is a var (not a
// const) so tests can point it at a local server.
var emojiCDN = "https://emoji.aranja.com/static/emoji-data/img-apple-160/"

// extractEmoji ensures the emoji PNG for the given codepoint sequence exists in
// ./images, downloading it if needed, and returns the local file path.
//
// The CDN is inconsistent about the trailing variation selector, so we try the
// full codepoint sequence first, then a version with a trailing '-fe0f'
// stripped.
func extractEmoji(unicode string) (string, error) {
	candidates := []string{unicode}
	if strings.HasSuffix(unicode, "-fe0f") {
		candidates = append(candidates, strings.TrimSuffix(unicode, "-fe0f"))
	}

	for _, candidate := range candidates {
		fileName := filepath.Join("images", candidate+".png")
		if _, err := os.Stat(fileName); err == nil {
			return fileName, nil
		}
		if err := download(emojiCDN+candidate+".png", fileName); err == nil {
			return fileName, nil
		}
		// download cleans up after itself, so there is nothing to remove here.
	}
	return "", fmt.Errorf("no emoji image found for U+%s", strings.ToUpper(unicode))
}

// download fetches url into dest, removing any partial file on failure so a
// later candidate isn't masked by an empty file.
func download(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(dest)
		return err
	}
	return f.Close()
}

func createImageFolder() error {
	return os.MkdirAll("images", 0o755)
}
