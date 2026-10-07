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

Versions are prereleases while the tool is young. Tag `vX.Y-beta`, two components,
starting at `v0.1-beta`, and bump `Y` for each release. Drop `-beta` when the API is
considered stable.

The leading `v` is GitHub convention and `release.yml` triggers on `tags: ["v*"]`. It
is not part of the version: inside the packages you get `0.1~beta` on Debian and
`0.1-beta` elsewhere. Two components, not three, so `0.2-beta` reads as "the second
release" rather than "less than one".

## Releasing

Tag `vX.Y-beta` and push the tag. GoReleaser builds the archives and packages, then
publishes them: a GitHub Release, the Homebrew tap, and the AUR package. Needs the
`HOMEBREW_TAP_TOKEN` repository secret for the tap.

The AUR publisher needs an `AUR_KEY` secret and a registered package. Until both
exist, `skip_upload: auto` in `.goreleaser.yaml` leaves the `PKGBUILD` in `dist/` for
review and skips the upload, which also covers every prerelease tag.

## Code

Every Go file starts with `// SPDX-License-Identifier: GPL-3.0-or-later`.
