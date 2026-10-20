// Package gosansmono holds Go Sans Mono, a monospaced sans-serif font made
// from the Go fonts, as TrueType data for a program to embed (as
// golang.org/x/image/font/gofont does for the Go fonts).
package gosansmono

import _ "embed"

// Regular is Go Sans Mono Regular, a TrueType font.
//
//go:embed fonts/GoSansMono-Regular.ttf
var Regular []byte

// Bold is Go Sans Mono Bold, a TrueType font.
//
//go:embed fonts/GoSansMono-Bold.ttf
var Bold []byte

// Version is the version of the fonts in this package (the fonts carry it as
// 1.002 in their name and head tables).
const Version = "1.0.2"

// Cell is the advance width of every glyph, in font units (2048 to the em),
// the same as Go Mono's.
const Cell = 1229
