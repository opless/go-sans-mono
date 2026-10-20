package gosansmono

import (
	"encoding/binary"
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

// TestNoHinting checks that the fonts carry no TrueType hinting. The Go
// sources are hinted, and the build moves and condenses their outlines, so
// the old instructions no longer fit: Windows runs them and drew a deformed
// Bold "o" (TREE-1616), whose outline lost two points. macOS ignores hinting.
func TestNoHinting(t *testing.T) {
	for name, data := range map[string][]byte{"Regular": Regular, "Bold": Bold} {
		tables := map[string][]byte{}
		for i := 0; i < int(binary.BigEndian.Uint16(data[4:])); i++ {
			r := data[12+16*i:]
			off, n := binary.BigEndian.Uint32(r[8:]), binary.BigEndian.Uint32(r[12:])
			tables[string(r[:4])] = data[off : off+n]
		}
		for _, tag := range []string{"fpgm", "prep", "cvt "} {
			if _, ok := tables[tag]; ok {
				t.Errorf("%s: has a %q table (TrueType hinting)", name, tag)
			}
		}
		// Each simple glyph keeps its own instructions after its contours.
		long := binary.BigEndian.Uint16(tables["head"][50:]) == 1
		loca, glyf := tables["loca"], tables["glyf"]
		hinted := 0
		for g := 0; g < int(binary.BigEndian.Uint16(tables["maxp"][4:])); g++ {
			var start, end uint32
			if long {
				start, end = binary.BigEndian.Uint32(loca[4*g:]), binary.BigEndian.Uint32(loca[4*g+4:])
			} else {
				start, end = 2*uint32(binary.BigEndian.Uint16(loca[2*g:])), 2*uint32(binary.BigEndian.Uint16(loca[2*g+2:]))
			}
			if start == end {
				continue
			}
			contours := int(int16(binary.BigEndian.Uint16(glyf[start:])))
			if contours < 0 {
				continue // composite; the build unlinks every reference
			}
			if binary.BigEndian.Uint16(glyf[start+10+2*uint32(contours):]) != 0 {
				hinted++
			}
		}
		if hinted > 0 {
			t.Errorf("%s: %d glyphs have TrueType instructions", name, hinted)
		}
	}
}
