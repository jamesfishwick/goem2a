# goem2a - Emoji to ASCII art (Go)

A Go port of [pyem2a](https://github.com/jamesfishwick/pyem2a), itself inspired
by [em2a](https://github.com/zmwangx/em2a) by
[zmwangx](https://github.com/zmwangx).

Compared to the Python version, the only external runtime dependency is `jp2a`.
Emoji-alias resolution ([`kyokomi/emoji`](https://github.com/kyokomi/emoji)) and
the PNG rasterizer ([`golang.org/x/image`](https://pkg.go.dev/golang.org/x/image),
which ships its own monospace font) are compiled into the binary, so there is no
Pillow/Menlo dependency.

## Dependencies

- [jp2a](https://github.com/Talinx/jp2a) at version 1.1.0 or newer (PNG support).
  On macOS: `brew install jp2a`.
- Go 1.21+ to build.

## Install / Build

```sh
go install github.com/jamesfishwick/goem2a@latest
# or, from a checkout:
go build -o goem2a .
```

## Usage

`goem2a` accepts both raw UTF-8 emojis and
[`:alias:`-styled escape sequences][1] used on GitHub and friends. Pass one or
more strings; an emoji may span several codepoints (e.g. a base char plus a
U+FE0F variation selector) and that is handled correctly.

```sh
goem2a ":smile:"            # terminal ASCII art
goem2a --color "🗃️"          # ANSI-colored art
goem2a --mirror ":wave:"    # horizontally mirrored
goem2a --png ":rocket:"     # also write <codepoint>_ascii.png (implies --color)
goem2a --bg '#1e1e2e' "🔥"   # set PNG background (implies --png)
```

| Flag | Effect |
| --- | --- |
| `--color` | Generate colored ASCII art. |
| `--mirror` | Mirror the result horizontally. |
| `--png` | Rasterize each result to `<codepoint>_ascii.png`. Implies `--color`. |
| `--bg <color>` | PNG background: `transparent` (RGBA), a name (`black`), or `#rrggbb`. Implies `--png`. Default white. |

jp2a's character ramp automatically follows the background's brightness (via
Rec. 601 luminance), so a dark `--bg` renders the emoji light-on-dark instead of
leaving it nearly blank.

## How it works

1. Resolve the argument to a literal emoji (alias lookup, or use as-is).
1. Join every rune's hex codepoint with `-` to form the CDN filename.
1. Download the Apple-style PNG from the emoji CDN into `./images` (cached),
   retrying without a trailing `-fe0f` since the CDN is inconsistent about it.
1. Shell out to `jp2a` to convert the PNG to (optionally colored) ASCII.
1. Optionally parse jp2a's ANSI SGR codes and redraw the art into a PNG with an
   embedded monospace font.

[1]: https://www.webpagefx.com/tools/emoji-cheat-sheet/
