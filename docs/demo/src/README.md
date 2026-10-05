# Demo GIF

`../demo.gif` is the animation at the top of the main README. This directory is
how it is made, kept in the repo so the animation can be edited and rebuilt
rather than re-recorded by hand.

```
python3 docs/demo/src/build_demo.py
```

Needs Pillow (`pip install pillow`). No shell commands are run, no network is
touched.

| File | What it is |
|---|---|
| `frames.py` | renders terminal frames and assembles the GIF |
| `build_demo.py` | the scenes: what is typed, what is answered |
| `demo-v1.gif` | the first version, kept for comparison |

## Editing a scene

`build_demo.py` is one function per scene. To change what a request says or what
the model answers, edit that scene and re-run. To change the whole look — font,
colours, chrome — edit the constants at the top of `frames.py`.

## What is real in the recording

Real: the prompts, the spinner, the `[y/N/e]` confirm, the `edit>` prompt, the
`Running:` line, the command output and the error message. All of it is verbatim
output of the binary, captured on 2026-10-05 against a local OpenAI-compatible
stub so that the build needs no API key and no network, and so it reproduces on
any machine.

Scripted: the model's answers in the first three scenes. A real model would
answer differently. The stub existed only so the demo could be built here.
