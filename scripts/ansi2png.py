#!/usr/bin/env python3
"""Render ANSI-coloured terminal lines to a PNG, via ImageMagick.

Draws each coloured run at an exact cell position using a monospace Nerd
Font, rather than letting an SVG renderer pick a font and flow the text.
"""
import sys, re, subprocess, unicodedata

NAMED = {
    30:"#4b5263",31:"#e06c75",32:"#98c379",33:"#e5c07b",34:"#61afef",35:"#c678dd",36:"#56b6c2",37:"#dcdfe4",
    90:"#5c6370",91:"#e06c75",92:"#98c379",93:"#e5c07b",94:"#61afef",95:"#c678dd",96:"#56b6c2",97:"#ffffff",
}
import os
FONT_DIR = os.environ.get("NERD_FONT_DIR", os.path.expanduser("~/Library/Fonts"))
FONT = f"{FONT_DIR}/JetBrainsMonoNerdFontMono-Regular.ttf"
FONT_BOLD = f"{FONT_DIR}/JetBrainsMonoNerdFontMono-Bold.ttf"
PT = 32
CELL_W = 19.025
LINE_H = 88      # pill height plus a full pill-height of air between rows
PAD_X = 44
PAD_Y = 40
RADIUS = 18
BG = "#0f1117"
FG = "#d8dee9"

sgr = re.compile(r"\x1b\[([0-9;]*)m")

def cells(text):
    """Terminal cell count: wide glyphs take two."""
    return sum(2 if unicodedata.east_asian_width(c) in ("W", "F") else 1 for c in text)

def parse(line):
    fg, bg, bold = None, None, False
    pos = 0
    for m in sgr.finditer(line):
        if m.start() > pos:
            yield line[pos:m.start()], fg, bg, bold
        codes = [int(c) for c in m.group(1).split(";") if c != ""] or [0]
        i = 0
        while i < len(codes):
            c = codes[i]
            if c == 0: fg, bg, bold = None, None, False
            elif c == 1: bold = True
            elif c == 22: bold = False
            elif c in NAMED: fg = NAMED[c]
            elif c - 10 in NAMED: bg = NAMED[c - 10]
            elif c in (38, 48) and i + 4 < len(codes) and codes[i+1] == 2:
                col = "#%02x%02x%02x" % (codes[i+2], codes[i+3], codes[i+4])
                if c == 38: fg = col
                else: bg = col
                i += 4
            i += 1
        pos = m.end()
    if pos < len(line):
        yield line[pos:], fg, bg, bold

def main():
    out_path = sys.argv[1]
    lines = [l.rstrip("\n") for l in sys.stdin]
    while lines and not lines[-1].strip():
        lines.pop()
    rows = [list(parse(l)) for l in lines]
    width_cells = max((sum(cells(t) for t, *_ in r) for r in rows), default=40)
    w = int(width_cells * CELL_W + 2 * PAD_X)
    h = int(len(rows) * LINE_H + 2 * PAD_Y)

    cmd = ["magick", "-size", f"{w}x{h}", "xc:none", "-fill", BG,
           "-draw", f"roundrectangle 0,0 {w-1},{h-1} {RADIUS},{RADIUS}",
           "-pointsize", str(PT)]
    for row, runs in enumerate(rows):
        col = 0
        # Each pill is drawn a fixed height, centred in its row, so the gap
        # between rows stays even however tall the line box is.
        row_top = PAD_Y + row * LINE_H
        pill_h = 46      # the pill itself; the rest of LINE_H is the gap
        top = row_top + (LINE_H - pill_h) / 2
        baseline = top + pill_h * 0.72
        for text, fg, bg, bold in runs:
            if not text:
                continue
            x = PAD_X + col * CELL_W
            span = cells(text)
            if bg:
                cmd += ["-fill", bg, "-draw",
                        f"rectangle {x:.1f},{top:.1f} {x + span*CELL_W:.1f},{top+pill_h:.1f}"]
            # -annotate takes the string as its own argument, so no quoting
            # rules can eat a glyph the way -draw text does.
            cmd += ["-font", FONT_BOLD if bold else FONT, "-fill", fg or FG,
                    "-annotate", f"+{x:.1f}+{baseline:.1f}", text]
            col += span
    cmd.append(out_path)
    subprocess.run(cmd, check=True)

main()
