package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("PNGDATA"))
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.png")
	if err := download(srv.URL, dest); err != nil {
		t.Fatalf("download: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "PNGDATA" {
		t.Errorf("downloaded content = %q, want PNGDATA", got)
	}
}

// A 404 must return an error and leave no partial file behind, so a later
// candidate (e.g. the -fe0f-stripped name) isn't masked by an empty file.
func TestDownloadNotFoundLeavesNoFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out.png")
	if err := download(srv.URL, dest); err == nil {
		t.Fatal("download of a 404 should return an error")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Errorf("expected no file at %s after a 404, stat err = %v", dest, err)
	}
}

// extractEmoji returns a cached file immediately without hitting the network.
func TestExtractEmojiCacheHit(t *testing.T) {
	chdirTemp(t)
	if err := createImageFolder(); err != nil {
		t.Fatalf("createImageFolder: %v", err)
	}
	want := filepath.Join("images", "1f5c3-fe0f.png")
	if err := os.WriteFile(want, []byte("cached"), 0o644); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	got, err := extractEmoji("1f5c3-fe0f")
	if err != nil {
		t.Fatalf("extractEmoji: %v", err)
	}
	if got != want {
		t.Errorf("extractEmoji cache hit = %q, want %q", got, want)
	}
}

// When the full codepoint PNG 404s, extractEmoji falls back to the
// -fe0f-stripped name and downloads that instead. This is the CDN-filename
// inconsistency the multi-codepoint fix handles.
func TestExtractEmojiFallsBackToStrippedVariationSelector(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only the stripped name exists; the full -fe0f sequence 404s.
		if r.URL.Path == "/1f5c3.png" {
			w.Write([]byte("stripped"))
			return
		}
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	origCDN := emojiCDN
	emojiCDN = srv.URL + "/"
	t.Cleanup(func() { emojiCDN = origCDN })

	chdirTemp(t)
	if err := createImageFolder(); err != nil {
		t.Fatalf("createImageFolder: %v", err)
	}

	got, err := extractEmoji("1f5c3-fe0f")
	if err != nil {
		t.Fatalf("extractEmoji: %v", err)
	}
	want := filepath.Join("images", "1f5c3.png") // downloaded under the stripped name
	if got != want {
		t.Errorf("extractEmoji fallback = %q, want %q", got, want)
	}
	if data, _ := os.ReadFile(got); string(data) != "stripped" {
		t.Errorf("downloaded content = %q, want stripped", data)
	}
}

// When neither candidate resolves, extractEmoji reports a clear error.
func TestExtractEmojiNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	origCDN := emojiCDN
	emojiCDN = srv.URL + "/"
	t.Cleanup(func() { emojiCDN = origCDN })

	chdirTemp(t)
	if err := createImageFolder(); err != nil {
		t.Fatalf("createImageFolder: %v", err)
	}

	if _, err := extractEmoji("1f5c3-fe0f"); err == nil {
		t.Fatal("extractEmoji should error when no candidate resolves")
	}
}

// chdirTemp switches into a fresh temp dir for the duration of the test, since
// extractEmoji/createImageFolder operate on a relative ./images path.
func chdirTemp(t *testing.T) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
}
