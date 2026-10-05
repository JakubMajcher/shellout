# Contributing

## Tests

`go test ./...` runs without network access or API keys. Shell integration tests
skip missing shells; set `SHELLOUT_REQUIRE_SHELLS=1` to make them fail instead (CI
does this).

## Manual smoke test before a release

With a real profile (OpenAI key or local Ollama), in a terminal:

1. `shellout list files sorted by size` shows a description and a command; `N` exits 1.
2. `y` runs it; `echo $?` shows the command's exit code.
3. `e` opens the command for editing; Enter runs the edited text.
4. `shellout` with no words shows `> `; `find files > 1G` works without quotes.
5. With `eval "$(shellout init zsh)"`: `sho go to /tmp` changes the directory and the
   up arrow shows the `cd` command.
6. A profile with `max_completion_tokens = 5` fails with
   `model ran out of tokens; raise the token limit in params`.

## Versioning

Versions are prereleases while the tool is young. Tag `vX.Y.Z-beta`, starting at
`v0.0.2-beta`, and bump `Y.Z` for each release. Drop `-beta` when the API is
considered stable.

## Releasing

Tag `vX.Y.Z-beta` and push the tag. GoReleaser builds the archives and packages,
updates the Homebrew tap and pushes the AUR package. Needs the `HOMEBREW_TAP_TOKEN`
and `AUR_KEY` repository secrets.

The `maintainer` field in `.goreleaser.yaml` must be filled in before the first
release; it currently holds a placeholder.

## Code

Every Go file starts with `// SPDX-License-Identifier: GPL-3.0-or-later`.
