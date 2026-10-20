# Go Sans Mono

A monospaced sans-serif font made from the [Go fonts](https://go.dev/blog/go-fonts)
by Bigelow & Holmes. Go Mono, the monospaced Go font, has slab serifs; Go
Sans Mono keeps the sans-serif Go design and puts it on Go Mono's grid.
It is made for TreeMendous and TreeMendous - Terminal (TREE-1391), to
replace JetBrains Mono, whose licence (OFL) is not MIT or BSD (TREE-1389).

**Version 1.0.2.** Regular and Bold, made entirely by the build script: no
glyph has been touched by hand yet (see Known work).

## How it is made

`build/build.py` (a FontForge script) does this for Regular and Bold:

1. Starts from Go Regular or Go Bold (`source/`, from `golang.org/x/image`
   v0.46.0, see `source/SOURCE.md`).
2. Takes these characters from Go Mono instead, because they are the same
   design already drawn for a monospaced cell:
   `i j l I 1 0` (feet and bars, and the slashed zero),
   `← ↑ → ↓ ↔ ↕`, `…`, `–` and `—`.
3. Condenses every glyph wider than the cell, keeping stem widths
   (FontForge `condenseExtend`); if that cannot make it fit, scales it.
   `condenseExtend` keeps only vertical stems, so diagonals and curves get
   thinner (`W` lost 40%). The glyphs listed in `LIGHT` are made heavier
   again in proportion to the squeeze (TREE-1617). They stay a little
   lighter than the Go fonts, so that a full cell looks as dark as the rest.
4. Centres every glyph in a cell of 1229 units (2048 to the em), Go Mono's
   advance width, so the metrics match Go Mono.
5. Removes kerning and the TrueType hinting, sets the fixed-pitch flags, and
   names the font. The Go sources are hinted, but their instructions fit the
   outlines before steps 3 and 4 (condensing can even drop points), and
   Windows, which runs them, drew a deformed Bold `o` (TREE-1616).

`build/check.py` then checks that every character is one cell wide (or zero,
for a combining mark), stays inside its cell, that the font says it is
fixed pitch, and that it carries no hinting.

## Build

Build the fonts you commit in the pinned container. It needs podman or
docker, nothing else:

```
make container  # builds fonts/ in build/Containerfile and runs the check
```

The container pins everything: the Debian image by digest, FontForge
20230101 and fontTools 4.57.0 by version from a dated snapshot.debian.org
archive, and the platform (linux/amd64, also on Apple Silicon). With the
`SOURCE_DATE_EPOCH` the Makefile sets, it gives the committed font files,
byte for byte.

To build with your own tools instead, you need FontForge (with Python) and
fontTools. On a Mac:

```
brew install fontforge fonttools
make            # builds fonts/ and runs the check
```

Another FontForge version can give slightly different outlines for a few
of the glyphs made heavier in step 3 (Homebrew's 20251009 changes 7 of
them, TREE-1618), so the files are not byte for byte the same.

## Use from Go

```
go get github.com/opless/go-sans-mono@v1.0.2
```

```go
import "github.com/opless/go-sans-mono"

face, err := opentype.Parse(gosansmono.Regular) // or gosansmono.Bold
```

`go test` checks that the embedded fonts parse, that the common
characters are one cell wide, that the fonts carry no hinting, and that
squeezed glyphs keep enough of the Go fonts' stroke weight.

## Known work

These glyphs need hand work, in both weights:

- `m w M W` are condensed to about 65 % and look cramped. Their weight is
  restored by the build (TREE-1617), but their shapes are still squeezed.
- `@ % &` are small. `Q`'s tail is long. `*` sits high.
- There is no Medium weight.
- Greek, Cyrillic and other non-Latin glyphs are condensed by the same
  script and have not been checked by eye.
- There is no hinting (the build drops the Go sources' hints, see step 5).
  macOS ignores hinting; Windows may render the font softer. ttfautohint
  could hint it again.

## Licence

BSD-3-Clause. See `LICENSE`, which also carries the Bigelow & Holmes notice
of the Go fonts, as their licence requires. The Go fonts' own licence text is
in `source/README-gofont`. Go Sans Mono is not made or endorsed by Google or
by Bigelow & Holmes.
