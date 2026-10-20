# Builds Go Sans Mono from the Go fonts in source/ (golang.org/x/image v0.46.0).
# "make container" builds in the pinned environment of build/Containerfile and
# gives the committed font files, byte for byte; use it for the fonts you
# commit. "make" uses the FontForge (with Python) and fontTools you have; on a
# Mac: brew install fontforge fonttools. Another FontForge version can give a
# few glyphs slightly different outlines.

FONTFORGE ?= fontforge
# fontTools' own Python, as Homebrew installs it; override if yours differs.
FONTTOOLS_PY ?= $(shell head -1 $$(command -v ttx) | sed 's/^\#!//')
# A fixed timestamp, so the same sources give the same font files.
export SOURCE_DATE_EPOCH = 1792022400

FONTS = fonts/GoSansMono-Regular.ttf fonts/GoSansMono-Bold.ttf

# podman or docker
CONTAINER ?= $(shell command -v podman >/dev/null 2>&1 && echo podman || echo docker)
IMAGE = go-sans-mono-build
# The checkout as the container engine sees it (on Windows, a C:/... path).
SRC ?= $(CURDIR)

.PHONY: all check clean container
all: $(FONTS) check

fonts/GoSansMono-Regular.ttf: build/build.py build/fixed_pitch.py source/Go-Regular.ttf source/Go-Mono.ttf
	$(FONTFORGE) -quiet -lang=py -script build/build.py source/Go-Regular.ttf source/Go-Mono.ttf $@ Regular
	$(FONTTOOLS_PY) build/fixed_pitch.py $@

fonts/GoSansMono-Bold.ttf: build/build.py build/fixed_pitch.py source/Go-Bold.ttf source/Go-Mono-Bold.ttf
	$(FONTFORGE) -quiet -lang=py -script build/build.py source/Go-Bold.ttf source/Go-Mono-Bold.ttf $@ Bold
	$(FONTTOOLS_PY) build/fixed_pitch.py $@

check: $(FONTS)
	$(FONTTOOLS_PY) build/check.py $(FONTS)

clean:
	rm -f $(FONTS)

container:
	$(CONTAINER) build --platform linux/amd64 -t $(IMAGE) -f build/Containerfile build
	$(CONTAINER) run --rm --platform linux/amd64 -v "$(SRC)":/src $(IMAGE) \
		sh -c 'rm -f $(FONTS) && make all'
