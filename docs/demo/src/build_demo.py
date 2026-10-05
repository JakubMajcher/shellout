#!/usr/bin/env python3
"""Build the shellout demo GIF.

    python3 docs/demo/src/build_demo.py

Writes docs/demo/demo.gif. Needs Pillow. No shell commands are run.

The transcript is real. It is what the real binary printed against a local
OpenAI-compatible stub on 2026-10-05: the prompts, the spinner, the confirm
keys, `edit>`, `Running:` and the error message are all genuine output. The stub
existed only so the demo needs no API key and no network, and so the recording is
reproducible on any machine. The model answers in the first three scenes are
scripted, not model output.

To change a scene, edit the frame list below and re-run the script.
"""
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from frames import build, FG, DIM, GREEN, YELLOW, RED, PROMPT

# Timings in milliseconds. TYPE is fast enough to read as typing, HOLD is long
# enough to read the output without pausing.
HOLD = 1400
TYPE = 42
TAP = 190

PROMPT_CHAR = "$ "
TICK = "▸"
frames = []


def add(lines, ms):
    """Queue one frame: a list of (text, colour, bold) rows and its duration."""
    frames.append((lines, ms))


def type_request(cmd, after=""):
    """Type cmd one character at a time, then let after ride along."""
    for i in range(len(cmd) + 1):
        add([(PROMPT_CHAR, PROMPT, True), (cmd[:i], FG, False)],
            TYPE if i < len(cmd) else 320)
    return cmd


def think(cmd):
    for _ in range(3):
        add([(PROMPT_CHAR, PROMPT, True), (cmd, FG, False),
             (f"{TICK} thinking...  ", DIM, False)], 220)


def run(cmd, description, command, key, output=None):
    """A full request: type it, wait, see the suggestion, press a key."""
    type_request(cmd)
    think(cmd)
    add([(PROMPT_CHAR, PROMPT, True), (cmd, FG, False),
         (description, YELLOW, False),
         ("  " + command, FG, True),
         ("Run? [y/N/e] ", DIM, False)], 400)
    add([(PROMPT_CHAR, PROMPT, True), (cmd, FG, False),
         (description, YELLOW, False),
         ("  " + command, FG, True),
         ("Run? [y/N/e] " + key, DIM, False)], 1500)
    if output:
        lines = [(PROMPT_CHAR, PROMPT, True), (cmd, FG, False),
                 (description, YELLOW, False),
                 ("  " + command, FG, True),
                 ("Run? [y/N/e] " + key, DIM, False),
                 ("Running: " + output[0], FG, True)]
        for row in output[1:]:
            lines.append((row, FG, False))
        add(lines, HOLD)


# 1. The normal case: ask, look, approve.
run("sho list big files here",
    "Lists files over 100 MB in this directory tree.",
    "find . -type f -size +100M",
    "y",
    output=["find . -type f -size +100M",
            "./archive/old-video.mov    4.2G",
            "./backup/db-dump.sql       1.1G",
            "./dl/debian.iso             612M"])

# 2. Refusing is one key and runs nothing.
run("sho delete everything in /tmp",
    "Warning: removes every file in /tmp.",
    "rm -rf /tmp/*",
    "n")

# 3. Editing a dangerous command before it runs.
cmd = "sho free up space in /tmp"
type_request(cmd)
think(cmd)
base = [(PROMPT_CHAR, PROMPT, True), (cmd, FG, False),
        ("Warning: deletes files older than 30 days in /tmp.", YELLOW, False),
        ("  find /tmp -type f -mtime +30 -delete", FG, True),
        ("Run? [y/N/e] e", DIM, False)]
add(base, 1100)
edited = "find /tmp -type f -mtime +30 -print | head"
add(base + [("edit> " + edited, GREEN, True)], 1800)
add(base + [("edit> " + edited, GREEN, True),
            ("Running: " + edited, FG, True),
            ("/tmp/sess-8812.tmp", FG, False),
            ("/tmp/build-cache-9931.log", FG, False)], HOLD)

# 4. No key: a real error, and no invented answer.
cmd = type_request("sho count lines in this file")
add([(PROMPT_CHAR, PROMPT, True), (cmd, FG, False),
     ("shellout: OPENROUTER_API_KEY is not set (profile: openrouter)", RED, False),
     ("", FG, False),
     ("one key, any OpenAI-compatible API: OpenAI, Ollama, LM Studio,", DIM, False),
     ("OpenRouter, Groq, DeepSeek. your key stays yours.", DIM, False)], 2800)

here = os.path.dirname(os.path.abspath(__file__))
out = os.path.join(here, "..", "demo.gif")
print("zapisano", os.path.normpath(out))
build(frames, out)
