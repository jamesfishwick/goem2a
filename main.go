package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"
)

// background captures how the rasterized PNG canvas is filled.
type background struct {
	transparent bool
	fill        color.RGBA
}

// options holds the parsed command-line flags.
type options struct {
	color          bool
	mirror         bool
	png            bool
	jp2aBackground string // "light" or "dark": jp2a's character ramp follows this
	bg             background
}

const usage = `Usage: goem2a <string> [<string> ...]

Each string can contain one or more raw UTF-8 emojis or :alias:-styled emoji escape sequences.
Optional the argument --color can be used to get colored emojis
Optional the argument --mirror can be used to get mirrored emojis
Optional the argument --png writes each result to a colored PNG image (implies --color)
Optional the argument --bg <color> sets the PNG background: "transparent", a name (black), or #rrggbb (implies --png)`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// run is the testable entry point: it returns an error instead of exiting so
// the orchestration can be exercised without process control.
func run(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Println(usage)
		return nil
	}

	opts, emojiArgs, err := parseArgs(args)
	if err != nil {
		return err
	}
	if err := createImageFolder(); err != nil {
		return err
	}

	type result struct {
		literal string
		ascii   string
	}
	var results []result

	for _, arg := range emojiArgs {
		literal, err := resolveEmoji(arg)
		if err != nil {
			return err
		}
		unicode := codepointHex(literal)

		fileName, err := extractEmoji(unicode)
		if err != nil {
			return err
		}

		ascii, err := emojiToASCII(fileName, opts)
		if err != nil {
			return err
		}
		results = append(results, result{literal, ascii})

		if opts.png {
			outPath := "./" + unicode + "_ascii.png"
			if err := renderASCIIToPNG(ascii, outPath, opts.bg); err != nil {
				return err
			}
			fmt.Printf("Wrote %s\n", outPath)
		}
	}

	for _, r := range results {
		fmt.Print("\n\n\n")
		fmt.Println(r.literal)
		fmt.Print("\n\n\n")
		fmt.Println(r.ascii)
		fmt.Print("\n\n\n")
	}
	return nil
}

// parseArgs consumes the flag arguments and returns the remaining emoji
// strings. It mirrors the Python script: --bg and --png both imply --color (the
// rasterizer needs jp2a's ANSI colors to preserve color).
func parseArgs(args []string) (options, []string, error) {
	opts := options{
		jp2aBackground: "light",
		bg:             background{fill: color.RGBA{255, 255, 255, 255}}, // white default
	}

	var rest []string
	var bgValue string
	haveBG := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case strings.HasPrefix(a, "--bg="):
			bgValue = strings.SplitN(a, "=", 2)[1]
			haveBG = true
		case a == "--bg":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("--bg requires a value (a color name, #rrggbb, or \"transparent\")")
			}
			bgValue = args[i+1]
			haveBG = true
			i++ // consume the value
		case a == "--png":
			opts.png = true
			opts.color = true
		case a == "--color":
			opts.color = true
		case a == "--mirror":
			opts.mirror = true
		default:
			rest = append(rest, a)
		}
	}

	if haveBG {
		if err := parseBackground(bgValue, &opts); err != nil {
			return opts, nil, err
		}
		opts.png = true   // a background only matters for the rasterized image
		opts.color = true // and the rasterizer needs jp2a's ANSI colors
	}

	if len(rest) == 0 {
		return opts, nil, fmt.Errorf("no emoji arguments given")
	}
	return opts, rest, nil
}

// parseBackground sets the canvas fill and chooses jp2a's light/dark ramp.
func parseBackground(value string, opts *options) error {
	switch strings.ToLower(value) {
	case "transparent", "none", "clear":
		opts.bg = background{transparent: true}
		// ambiguous luminance; keep the default light ramp
		return nil
	}
	c, err := parseColor(value)
	if err != nil {
		return err
	}
	opts.bg = background{fill: c}
	// a dark canvas wants jp2a's dark ramp
	if luminance(c) < 0.5 {
		opts.jp2aBackground = "dark"
	} else {
		opts.jp2aBackground = "light"
	}
	return nil
}
