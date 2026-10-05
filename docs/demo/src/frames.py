#!/usr/bin/env python3
"""Draw terminal frames and assemble them into an animated GIF.

Used by build_demo.py. Each frame is a list of (text, colour, bold) rows; this
module renders them with a monospace font on a fixed canvas and writes a GIF
where every frame keeps its own duration.

Two decisions worth keeping if you edit this:

- One canvas size for all frames, taken from the widest and tallest frame. Frames
  have different line counts, so sizing each frame to its own content makes the
  GIF jump vertically between scenes.
- Assembled with Pillow, not with a two-pass ffmpeg palette. The ffmpeg route
  collapsed the per-frame durations, and the timings are what make typing read
  as typing rather than as a slideshow.
"""
from PIL import Image, ImageDraw, ImageFont

FONT = "/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"
FONT_BOLD = "/usr/share/fonts/truetype/dejavu/DejaVuSansMono-Bold.ttf"

SIZE = 26
PAD = 24
CHROME = 34
LINE_H = SIZE + 10

BG = (24, 26, 33)
CHROME_BG = (38, 41, 50)
FG = (222, 226, 233)
DIM = (128, 134, 148)
GREEN = (140, 210, 140)
YELLOW = (232, 200, 120)
RED = (240, 130, 130)
PROMPT = (126, 208, 166)

_font = ImageFont.truetype(FONT, SIZE)
_font_bold = ImageFont.truetype(FONT_BOLD, SIZE)
_char_w = _font.getbbox("M")[2] - _font.getbbox("M")[0]


def measure(lines):
    cols = max((len(text) for text, *_ in lines), default=1)
    return cols * _char_w + PAD * 2, len(lines) * LINE_H + PAD * 2 + CHROME


def render(lines, width, height):
    img = Image.new("RGB", (width, height), BG)
    d = ImageDraw.Draw(img)

    # Window chrome, so it reads as a terminal and not as a code screenshot.
    d.rectangle([0, 0, width, CHROME], fill=CHROME_BG)
    for i, c in enumerate([(255, 95, 86), (255, 189, 46), (39, 201, 63)]):
        d.ellipse([16 + i * 20, 12, 26 + i * 20, 22], fill=c)
    label = "shellout"
    label_w = _font_bold.getbbox(label)[2] - _font_bold.getbbox(label)[0]
    d.text(((width - label_w) // 2, 9), label, font=_font_bold, fill=DIM)

    y = PAD + CHROME
    for text, color, bold in lines:
        if text:
            d.text((PAD, y), text, font=_font_bold if bold else _font, fill=color)
        y += LINE_H

    # A block cursor on the last non-empty line. Without it the demo looks frozen
    # on frames where nothing new is being typed.
    for text, _c, _b in reversed(lines):
        if text:
            x = PAD + len(text) * _char_w
            d.rectangle([x, y - LINE_H + 6, x + _char_w - 3, y - 3], fill=GREEN)
            break
    return img


def build(frames, out, scale=1):
    """frames: list of (rows, duration_ms). Returns the output path."""
    width = max(measure(rows)[0] for rows, _ in frames)
    height = max(measure(rows)[1] for rows, _ in frames)
    images = [render(rows, width, height) for rows, _ in frames]
    durations = [ms for _, ms in frames]

    if scale != 1:
        images = [im.resize((im.width * scale, im.height * scale), Image.LANCZOS)
                  for im in images]

    print(f"{len(images)} frames, {width}x{height}, "
          f"{sum(durations) / 1000:.1f}s")
    images[0].save(out, save_all=True, append_images=images[1:],
                   duration=durations, loop=0, optimize=True, disposal=2)
    return out
