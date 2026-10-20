package gosansmono

import (
	"encoding/binary"
	"math"
	"os"
	"sort"
	"testing"
	"unicode"

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

// TestStrokeWeight checks that squeezing wide glyphs into the cell did not
// leave their strokes far lighter than the Go fonts' (TREE-1617: W lost 40%
// of its weight). It measures each stroke across, from where two horizontal
// lines cut the outline, and compares it with the same stroke in the source.
// Squeezed glyphs are meant to be a little lighter than the source, so that
// they look as dark as the rest; @ % ¼ ¾ are left out (see build.py), and
// Q @ % & come from Go Mono (TestFromGoMono).
func TestStrokeWeight(t *testing.T) {
	for _, w := range []string{"Regular", "Bold"} {
		src, err := os.ReadFile("source/Go-" + w + ".ttf")
		if err != nil {
			t.Fatal(err)
		}
		out := map[string][]byte{"Regular": Regular, "Bold": Bold}[w]
		for _, r := range "WwMmYKVXŒœÆæØΜΦΨΩωπЉљШЩщФф" {
			a, b := strokes(t, src, r), strokes(t, out, r)
			if len(a) == 0 || len(a) != len(b) {
				continue // the strokes do not line up; nothing to compare
			}
			for i := range a {
				if a[i] < 60 {
					continue
				}
				if ratio := b[i] / a[i]; ratio < 0.7 || ratio > 1.25 {
					t.Errorf("%s %q: stroke %d is %.0f units across, %.0f in the Go font (%.2f)", w, r, i, b[i], a[i], ratio)
				}
			}
		}
	}
}

// strokes returns the thickness of each stroke that crosses the glyph at 40%
// and 50% of the x-height (lowercase) or cap height, measured at right
// angles to the stroke, in font units.
func strokes(t *testing.T, data []byte, r rune) []float64 {
	f, err := sfnt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf sfnt.Buffer
	g, err := f.GlyphIndex(&buf, r)
	if err != nil || g == 0 {
		return nil
	}
	segs, err := f.LoadGlyph(&buf, g, fixed.I(int(f.UnitsPerEm())), nil)
	if err != nil {
		t.Fatal(err)
	}
	// Flatten the outline into line segments (y grows downwards).
	type pt struct{ x, y float64 }
	p := func(q fixed.Point26_6) pt { return pt{float64(q.X) / 64, float64(q.Y) / 64} }
	var lines [][2]pt
	var start, cur pt
	for _, s := range segs {
		switch s.Op {
		case sfnt.SegmentOpMoveTo:
			if cur != start {
				lines = append(lines, [2]pt{cur, start})
			}
			start, cur = p(s.Args[0]), p(s.Args[0])
		case sfnt.SegmentOpLineTo:
			lines = append(lines, [2]pt{cur, p(s.Args[0])})
			cur = p(s.Args[0])
		case sfnt.SegmentOpQuadTo, sfnt.SegmentOpCubeTo:
			c := []pt{cur, p(s.Args[0]), p(s.Args[1]), p(s.Args[2])}[:map[sfnt.SegmentOp]int{sfnt.SegmentOpQuadTo: 3, sfnt.SegmentOpCubeTo: 4}[s.Op]]
			prev := cur
			for k := 1; k <= 16; k++ {
				u := float64(k) / 16
				q := append([]pt(nil), c...)
				for n := len(q) - 1; n > 0; n-- { // de Casteljau
					for j := 0; j < n; j++ {
						q[j] = pt{q[j].x + (q[j+1].x-q[j].x)*u, q[j].y + (q[j+1].y-q[j].y)*u}
					}
				}
				lines = append(lines, [2]pt{prev, q[0]})
				prev = q[0]
			}
			cur = c[len(c)-1]
		}
	}
	if cur != start {
		lines = append(lines, [2]pt{cur, start})
	}
	// Runs of ink along y: pairs of crossings, by the non-zero winding rule.
	runs := func(y float64) [][2]float64 {
		type cross struct {
			x   float64
			dir int
		}
		var cs []cross
		for _, l := range lines {
			a, b := l[0], l[1]
			if (a.y <= y) == (b.y <= y) {
				continue
			}
			dir := 1
			if b.y < a.y {
				dir = -1
			}
			cs = append(cs, cross{a.x + (y-a.y)*(b.x-a.x)/(b.y-a.y), dir})
		}
		sort.Slice(cs, func(i, j int) bool { return cs[i].x < cs[j].x })
		var out [][2]float64
		wind := 0
		for _, c := range cs {
			if wind == 0 {
				out = append(out, [2]float64{c.x, c.x})
			}
			wind += c.dir
			if wind == 0 {
				out[len(out)-1][1] = c.x
			}
		}
		return out
	}
	top := 1400.0
	if unicode.IsLower(r) {
		top = 1000
	}
	ya, yb := -top*0.4, -top*0.5
	ra, rb := runs(ya), runs(yb)
	if len(ra) != len(rb) {
		return nil
	}
	var th []float64
	for i := range ra {
		dx := (rb[i][0]+rb[i][1])/2 - (ra[i][0]+ra[i][1])/2
		sin := (ya - yb) / math.Hypot(dx, ya-yb)
		th = append(th, ((ra[i][1]-ra[i][0])+(rb[i][1]-rb[i][0]))/2*sin)
	}
	return th
}

// TestFromGoMono checks that Q @ % & are Go Mono's glyphs as drawn, not
// squeezed copies of Go's wider ones (TREE-1619). Bold % is fitted into the
// cell, which moves its points by up to 2 units.
func TestFromGoMono(t *testing.T) {
	for w, out := range map[string][]byte{"Regular": Regular, "Bold": Bold} {
		src, err := os.ReadFile(map[string]string{"Regular": "source/Go-Mono.ttf", "Bold": "source/Go-Mono-Bold.ttf"}[w])
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range "Q@%&" {
			a, b := outline(t, src, r), outline(t, out, r)
			if len(a) != len(b) {
				t.Errorf("%s %q: %d points, Go Mono has %d", w, r, len(b), len(a))
				continue
			}
			for i := range a {
				if d := a[i].Sub(b[i]); d.X > 2*64 || d.X < -2*64 || d.Y > 2*64 || d.Y < -2*64 {
					t.Errorf("%s %q: point %d is %v, Go Mono has %v", w, r, i, b[i], a[i])
					break
				}
			}
		}
	}
}

// TestAsterisk checks that * is centred on the middle of +, as an operator
// in code, not raised like a footnote mark (TREE-1619).
func TestAsterisk(t *testing.T) {
	for w, data := range map[string][]byte{"Regular": Regular, "Bold": Bold} {
		f, err := sfnt.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		var buf sfnt.Buffer
		mid := func(r rune) fixed.Int26_6 {
			g, err := f.GlyphIndex(&buf, r)
			if err != nil {
				t.Fatal(err)
			}
			b, _, err := f.GlyphBounds(&buf, g, fixed.I(int(f.UnitsPerEm())), 0)
			if err != nil {
				t.Fatal(err)
			}
			return (b.Min.Y + b.Max.Y) / 2
		}
		if d := mid('*') - mid('+'); d > fixed.I(2) || d < -fixed.I(2) {
			t.Errorf("%s: the middle of * is %v units from the middle of +", w, d)
		}
	}
}

// outline returns every point of a glyph's outline, in font units (26.6).
func outline(t *testing.T, data []byte, r rune) []fixed.Point26_6 {
	f, err := sfnt.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	var buf sfnt.Buffer
	g, err := f.GlyphIndex(&buf, r)
	if err != nil || g == 0 {
		t.Fatalf("no glyph for %q", r)
	}
	segs, err := f.LoadGlyph(&buf, g, fixed.I(int(f.UnitsPerEm())), nil)
	if err != nil {
		t.Fatal(err)
	}
	var pts []fixed.Point26_6
	for _, s := range segs {
		n := map[sfnt.SegmentOp]int{sfnt.SegmentOpMoveTo: 1, sfnt.SegmentOpLineTo: 1, sfnt.SegmentOpQuadTo: 2, sfnt.SegmentOpCubeTo: 3}[s.Op]
		pts = append(pts, s.Args[:n]...)
	}
	return pts
}
