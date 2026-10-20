# check.py <ttf>...: fail unless every mapped glyph is one cell wide (Go
# Mono's 1229 units, or 0 for a combining mark), lies inside its cell, the
# font says it is fixed pitch, and the font carries no TrueType hinting
# (the Go sources' hints do not fit the changed outlines; TREE-1616). Run with fontTools' Python.
import sys

from fontTools.pens.boundsPen import BoundsPen
from fontTools.ttLib import TTFont

CELL = 1229
failed = False
for path in sys.argv[1:]:
    font = TTFont(path)
    glyphs = font.getGlyphSet()
    widths = font["hmtx"].metrics
    problems = []
    cmap = font.getBestCmap()
    for code, name in sorted(cmap.items()):
        width = widths[name][0]
        if width not in (0, CELL):
            problems.append(f"U+{code:04X} {name}: width {width}")
            continue
        pen = BoundsPen(glyphs)
        glyphs[name].draw(pen)
        if width and pen.bounds and (pen.bounds[0] < 0 or pen.bounds[2] > CELL):
            problems.append(f"U+{code:04X} {name}: outside the cell {pen.bounds[0]:.0f}..{pen.bounds[2]:.0f}")
    if not font["post"].isFixedPitch:
        problems.append("post.isFixedPitch is not set")
    for tag in ("fpgm", "prep", "cvt "):
        if tag in font:
            problems.append(f"has a {tag!r} table (TrueType hinting)")
    hinted = [n for n in font.getGlyphOrder()
              if hasattr(font["glyf"][n], "program") and font["glyf"][n].program.getBytecode()]
    if hinted:
        problems.append(f"{len(hinted)} glyphs have TrueType instructions, e.g. {hinted[0]}")
    # Every name that identifies the font must be Go Sans Mono's, not Go's.
    version = f"Version {font['head'].fontRevision:.3f}"
    for rec in font["name"].names:
        text = rec.toUnicode()
        if rec.nameID == 5 and text != version:
            problems.append(f"name {rec.nameID} (platform {rec.platformID}) is {text!r}, want {version!r}")
        if rec.nameID in (1, 3, 4, 6, 18) and "Go Sans Mono" not in text and "GoSansMono" not in text:
            problems.append(f"name {rec.nameID} (platform {rec.platformID}) is {text!r}")
    print(f"{path}: {len(cmap)} characters, {len(problems)} problems")
    for p in problems[:20]:
        print("  " + p)
    failed = failed or bool(problems)
sys.exit(1 if failed else 0)
