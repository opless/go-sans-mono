# build.py <go sans ttf> <go mono ttf> <out ttf> <style>: Go Sans Mono, a
# monospaced sans-serif font made from the Go fonts (BSD-3, Bigelow & Holmes).
# Run by the Makefile as "fontforge -lang=py -script build/build.py ...".
import sys
import fontforge
import psMat

src, mono_src, out, style = sys.argv[1:5]
VERSION = "1.003"    # the font version; keep in step with Version in gosansmono.go
TARGET = 1229        # Go Mono's advance width, so the metrics match Go Mono
MIN_SB = 70          # smallest side bearing for a condensed glyph
# Narrow letters that look lost when merely centred: monospaced sans fonts
# give them feet or bars; Go Mono's are the same design, so take those.
FROM_MONO = set(map(ord, "iljI10")) | set(map(ord, "\u2190\u2191\u2192\u2193\u2194\u2195\u2026\u2014\u2013"))
# Wide glyphs that came out cramped when squeezed (TREE-1619): the bowl of Q
# got narrower than O and its tail long, and @ % & small. Go Mono draws them
# for this cell, so take them with Go Mono's own spacing, unsqueezed.
MONO_AS_IS = set(map(ord, "Q@%&"))

# condenseExtend keeps only the vertical stems it finds; it squeezes
# diagonals and curves like counters, so these glyphs came out up to half as
# heavy as the Go sources (TREE-1617; found by measuring every glyph's
# strokes against the source, as TestStrokeWeight does). They are made heavier
# again in proportion to how much they were squeezed. Regular m keeps its
# stems, and @ % ¼ ¾ go too heavy, so they are not on the lists.
LIGHT = {
    "Regular": set(map(ord, "WŴwŵMŒœÆæǼǽØǾYÝŶŸ"
                            "ΜΦΨΩωώΌΎΏπ"
                            "ЉљЊњЂЋМШЩщФф")),
    "Bold": set(map(ord, "WŴwŵMmŒœÆæǼǽØǾYÝŶŸKĶVXOÒÓÔÕÖŌŎŐǑGĜĞĠĢÐĐďĲ"
                         "ΜΦΨΩωώΌΎΏπΘΟΔΚΧχψσφ"
                         "ЉљЊњЂЋМмШшЩщФфОДыъЎУХЮ")),
}
K = {"Regular": 0.4, "Bold": 0.4}
STEM = {"Regular": 197, "Bold": 296}   # the stem of n in Go Regular and Go Bold
# changeWeight needs cubic outlines, but converting the whole font changes
# hundreds of glyphs on the way back to TrueType, so only these go through a
# cubic scratch font.
scratch = fontforge.font()
scratch.is_quadratic = False


def embolden(g, amount):
    s = scratch.createChar(-1, "tmp")
    s.clear()
    s.foreground = g.foreground
    s.width = g.width
    # "squish" keeps the glyph as wide as it is; the strokes take the room.
    s.changeWeight(amount, "auto", 0, 0, "squish")
    s.removeOverlap()
    g.foreground = s.foreground
    g.round()


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
    if u in MONO_AS_IS and u in mono:
        mono.selection.select(u)
        mono.copy()
        f.selection.select(u)
        f.paste()
        g = f[u]
        g.width = TARGET
        # Go Mono Bold's % is 2 units wider than the cell; fit it, nothing more.
        xmin, ymin, xmax, ymax = g.boundingBox()
        if xmin < 0 or xmax > TARGET:
            g.transform(psMat.scale(min(1, TARGET / (xmax - xmin)), 1))
            xmin, ymin, xmax, ymax = g.boundingBox()
            g.transform(psMat.translate((TARGET - (xmax - xmin)) / 2 - xmin, 0))
            g.round()
            g.width = TARGET
        continue
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
        if u in LIGHT[style]:
            embolden(g, K[style] * (1 - (xmax - xmin) / bw) * STEM[style])
            xmin, ymin, xmax, ymax = g.boundingBox()
        bw = xmax - xmin
    # Centre in the cell.
    g.transform(psMat.translate((TARGET - bw) / 2 - xmin, 0))
    g.width = TARGET

# Go's asterisk sits high, like a footnote mark; in code it is an operator
# (a*b, *p), so centre it on the middle of + as code fonts do (TREE-1619).
_, py0, _, py1 = f["plus"].boundingBox()
_, ay0, _, ay1 = f["asterisk"].boundingBox()
f["asterisk"].transform(psMat.translate(0, round((py0 + py1) / 2 - (ay0 + ay1) / 2)))

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
