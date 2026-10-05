# SH»OUT

`shellout` turns one English sentence into one shell command, shows it, and runs it
after you confirm. It works with any OpenAI-compatible API: OpenAI, Ollama,
LM Studio, OpenRouter, Groq, vLLM. `sho` is a short alias for the same program.

```
$ sho find files bigger than 1 GB here
Lists files larger than 1 GB under the current directory.
  find . -type f -size +1G
Run? [y/N/e]
```

`y` runs it, `e` lets you edit it first, anything else cancels.

## Install

- Homebrew (macOS, Linux): `brew install JakubMajcher/tap/shellout`
- Arch Linux (AUR): `yay -S shellout-bin`
- Debian/Ubuntu: download the `.deb` from Releases, then `sudo apt install ./shellout_*.deb`
- Fedora/openSUSE: download the `.rpm`, then `sudo dnf install ./shellout-*.rpm`
- Alpine: download the `.apk`, then `sudo apk add --allow-untrusted ./shellout_*.apk`
- Any Linux or macOS: download the archive from Releases and put `shellout` on your PATH.
- Go: `go install github.com/JakubMajcher/shellout/cmd/shellout@latest`
  (installs only `shellout`; add `alias sho=shellout` yourself)

## Configure

The first run creates `~/.config/shellout/config.toml` (or under `$XDG_CONFIG_HOME`,
or at `$SHELLOUT_CONFIG`) and stops, so you can edit it before the first request.

```toml
default = "openai"

[profiles.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-5-mini"
api_key_env = "OPENAI_API_KEY"
timeout = "60s"

[profiles.openai.params]
reasoning_effort = "minimal"
max_completion_tokens = 1000

# A local model through Ollama, no API key needed.
# [profiles.local]
# base_url = "http://localhost:11434/v1"
# model = "qwen2.5-coder:7b"
```

- Profiles are named. `default` picks the one used without `-p`; `SHELLOUT_PROFILE`
  and then `-p NAME` override it, in that order.
- `api_key_env` holds the **name** of the environment variable that holds the key,
  never the key itself, so the file is safe to keep in dotfiles. Leave it out for
  local servers that need no authentication.
- `params` is copied into the request body as-is, and shellout sends no other
  optional parameter. Use it for model settings and to cap cost
  (`max_completion_tokens`, `reasoning_effort`). Parameter names differ between
  providers; use the ones yours accepts. `params` may not set `model`, `messages`,
  `response_format` or `stream`, because shellout sets those itself.
- **OpenRouter**: add `provider = { require_parameters = true }` to `params`,
  or OpenRouter may route the request to a provider that ignores the JSON schema.
- The model must support structured outputs (`response_format: json_schema`).
  shellout validates every answer itself and retries once, but it cannot invent a
  command the endpoint never sent.

## Quoting

The shell parses your words before shellout sees them, so `sho find files > 1G`
creates a file named `1G`. Run `sho` with no words and type the request at the `>`
prompt instead.

## Shell integration (optional)

Without it, the command runs in a child process: `cd` and `export` have no effect,
your aliases and functions from `.zshrc` are not available, and history shows
`sho ...` instead of the command. With it, the approved command runs in your current
shell and goes into history.

- zsh: `eval "$(shellout init zsh)"` in `~/.zshrc`
- bash: `eval "$(shellout init bash)"` in `~/.bashrc`
- fish: `shellout init fish | source` in `~/.config/fish/config.fish`
  (history needs fish 4.0 or newer; on fish 3.x the command still runs, it just is
  not added to history)

## Privacy

shellout sends the provider your request, OS, CPU architecture, distribution name,
shell name and whether core utilities are GNU, BSD or BusyBox. Nothing else: no file
names, no paths, no current directory, no environment variables.

## License

GPL-3.0-or-later.
