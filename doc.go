// Copyright (C) 2026 James Fishwick
//
// This program is free software; you can redistribute it and/or modify it under
// the terms of the GNU General Public License version 2 as published by the Free
// Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

// Command goem2a renders emojis as ASCII art.
//
// It accepts raw UTF-8 emojis or :alias:-styled escape sequences, downloads the
// matching Apple-style emoji image from a CDN (cached under ./images), and
// converts it to ASCII art with the jp2a binary. Results can optionally be
// colored, mirrored, and rasterized back to a PNG using an embedded monospace
// font.
//
// Usage:
//
//	goem2a [flags] <string> [<string> ...]
//
// Each string may contain one or more emojis; a single emoji can span several
// codepoints (e.g. a base character plus a U+FE0F variation selector) and is
// handled as a unit.
//
// Flags:
//
//	--color        generate colored ASCII art
//	--mirror       mirror the result horizontally
//	--png          also write <codepoint>_ascii.png (implies --color)
//	--bg <color>   PNG background: "transparent", a name (black), or #rrggbb (implies --png)
//
// jp2a's character ramp follows the background brightness, so a dark --bg
// renders the emoji light-on-dark rather than nearly blank.
//
// goem2a requires the jp2a binary (version 1.1.0 or newer) on PATH.
package main
