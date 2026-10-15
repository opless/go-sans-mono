package gosansmono

import (
	"testing"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// TestMonospaced parses each embedded font and checks that the characters
// a terminal and a file list show most are all one cell wide.
func TestMonospaced(t *testing.T) {
	for name, data := range map[string][]byte{"Regular": Regular, "Bold": Bold} {
		f, err := sfnt.Parse(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		var buf sfnt.Buffer
		ppem := fixed.I(int(f.UnitsPerEm()))
		for _, r := range " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~éß€←→↑↓…—" {
			g, err := f.GlyphIndex(&buf, r)
			if err != nil || g == 0 {
				t.Errorf("%s: no glyph for %q", name, r)
				continue
			}
			adv, err := f.GlyphAdvance(&buf, g, ppem, 0)
			if err != nil {
				t.Errorf("%s: %q: %v", name, r, err)
				continue
			}
			if adv != fixed.I(Cell) {
				t.Errorf("%s: %q is %v wide, want %d", name, r, adv, Cell)
			}
		}
	}
}
