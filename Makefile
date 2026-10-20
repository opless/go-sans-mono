# Builds Go Sans Mono from the Go fonts in source/ (golang.org/x/image v0.46.0).
# Needs FontForge (with Python) and fontTools; on a Mac: brew install fontforge fonttools.

FONTFORGE ?= fontforge
# fontTools' own Python, as Homebrew installs it; override if yours differs.
FONTTOOLS_PY ?= $(shell head -1 $$(command -v ttx) | sed 's/^\#!//')
# A fixed timestamp, so the same sources give the same font files.
export SOURCE_DATE_EPOCH = 1792022400

FONTS = fonts/GoSansMono-Regular.ttf fonts/GoSansMono-Bold.ttf

.PHONY: all check clean
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
