# fixed_pitch.py <ttf>: set post.isFixedPitch. FontForge sets it itself only
# in some versions (Homebrew's did, Debian's 20230101 does not, perhaps
# because of the zero-width combining marks), so the build sets it here.
# Run with fontTools' Python.
import os
import sys

from fontTools.ttLib import TTFont

path = sys.argv[1]
font = TTFont(path)
if not font["post"].isFixedPitch:
    font["post"].isFixedPitch = 1
    # Keep head.modified at SOURCE_DATE_EPOCH, so the file stays reproducible.
    font.recalcTimestamp = False
    font.save(path + ".tmp")
    os.replace(path + ".tmp", path)
