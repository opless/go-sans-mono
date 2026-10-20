# build.py <go sans ttf> <go mono ttf> <out ttf> <style>: Go Sans Mono, a
# monospaced sans-serif font made from the Go fonts (BSD-3, Bigelow & Holmes).
# Run by the Makefile as "fontforge -lang=py -script build/build.py ...".
import sys
import fontforge
import psMat

src, mono_src, out, style = sys.argv[1:5]
VERSION = "1.001"    # the font version; keep in step with Version in gosansmono.go
TARGET = 1229        # Go Mono's advance width, so the metrics match Go Mono
MIN_SB = 70          # smallest side bearing for a condensed glyph
# Narrow letters that look lost when merely centred: monospaced sans fonts
# give them feet or bars; Go Mono's are the same design, so take those.
FROM_MONO = set(map(ord, "iljI10")) | set(map(ord, "\u2190\u2191\u2192\u2193\u2194\u2195\u2026\u2014\u2013"))

f = fontforge.open(src)
mono = fontforge.open(mono_src)

# Kerning makes no sense in a monospaced font.
for l in list(f.gpos_lookups):
    f.removeLookup(l)

for g in list(f.glyphs()):
    if g.references:
        g.unlinkRef()

for g in list(f.glyphs()):
    u = g.unicode
    if u in FROM_MONO and u in mono:
        mono.selection.select(u)
        mono.copy()
        f.selection.select(u)
        f.paste()
        g = f[u]
    if g.width == 0:             # combining marks keep zero width
        continue
    xmin, ymin, xmax, ymax = g.boundingBox()
    bw = xmax - xmin
    if bw > TARGET - 2 * MIN_SB:
        # Condense the counters, keeping the stems, until the glyph fits.
        factor = (TARGET - 2 * MIN_SB) / bw
        for _ in range(6):
            trial = g.foreground.dup()
            g.condenseExtend(factor, 0)
            xmin, ymin, xmax, ymax = g.boundingBox()
            if xmax - xmin <= TARGET - 2 * MIN_SB:
                break
            g.foreground = trial
            factor *= 0.95
        else:
            # condenseExtend could not make it fit: scale it plainly.
            xmin, ymin, xmax, ymax = g.boundingBox()
            g.transform(psMat.scale((TARGET - 2 * MIN_SB) / (xmax - xmin), 1))
        xmin, ymin, xmax, ymax = g.boundingBox()
        bw = xmax - xmin
    # Centre in the cell.
    g.transform(psMat.translate((TARGET - bw) / 2 - xmin, 0))
    g.width = TARGET

fam = "Go Sans Mono"
f.familyname = fam
f.fontname = "GoSansMono-" + style
f.fullname = fam + " " + style
f.version = VERSION
f.sfntRevision = float(VERSION)
f.copyright = (f.copyright or "") + " Derived for TreeMendous (TREE-1391): monospaced, from the Go fonts."
f.os2_panose = (2, 11, 5, 9, 2, 2, 3, 2, 2, 4) if style == "Regular" else (2, 11, 8, 9, 2, 2, 3, 2, 2, 4)
# Replace the Go font's own names, not add to them: some FontForge versions
# (Debian's 20230101) keep the source's Version, and every version kept its
# UniqueID and Compatible Full ("Go Bold"), which can clash with an installed
# Go font.
names = {
    "Family": fam,
    "SubFamily": style,
    "UniqueID": "GoSansMono-" + style + " " + VERSION,
    "Fullname": fam + " " + style,
    "PostScriptName": "GoSansMono-" + style,
    "Version": "Version " + VERSION,
    "Compatible Full": fam + " " + style,
}
f.sfnt_names = tuple(n for n in f.sfnt_names if n[1] not in names)
for strid, value in names.items():
    f.appendSFNTName("English (US)", strid, value)
# The Go sources are hinted, but their instructions were written for the
# outlines before they were moved and condensed (condensing can even drop
# points), so Windows, which runs them, drew a deformed Bold "o" (TREE-1616).
# Ship the font unhinted, as macOS draws it anyway.
for g in f.glyphs():
    g.ttinstrs = b""
for tag in ("fpgm", "prep", "cvt "):
    f.setTableData(tag, None)
f.generate(out, flags=("opentype", "omit-instructions"))
print("wrote", out)
