# check.py <ttf>...: fail unless every mapped glyph is one cell wide (Go
# Mono's 1229 units, or 0 for a combining mark), lies inside its cell, and
# the font says it is fixed pitch. Run with fontTools' Python.
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
    print(f"{path}: {len(cmap)} characters, {len(problems)} problems")
    for p in problems[:20]:
        print("  " + p)
    failed = failed or bool(problems)
sys.exit(1 if failed else 0)
