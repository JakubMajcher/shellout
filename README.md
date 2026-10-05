# SH»OUT

**Say what you want. See the command. Press y.**

![demo](docs/demo/demo.gif)

`shellout` turns one English sentence into one shell command, shows it, and runs it
after you confirm. It works with any OpenAI-compatible API: OpenAI, Ollama,
LM Studio, OpenRouter, Groq, DeepSeek, vLLM. `sho` is a short alias for the same
program.

```
$ sho find files bigger than 1 GB here
Lists files larger than 1 GB under the current directory.
  find . -type f -size +1G
Run? [y/N/e]
```

`y` runs it, `e` lets you edit it first, anything else cancels.

Nothing runs without your keypress. No agent loop, no reading your output, no
files or paths sent anywhere — the model gets your request, your OS, your shell
and nothing else.

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

The file it writes already has profiles for the providers that speak the
OpenAI-compatible API, ready to use:

| Profile | Needs | Notes |
|---|---|---|
| `openrouter` | `OPENROUTER_API_KEY` | one key, many models; change `model`, not the profile |
| `openai` | `OPENAI_API_KEY` | |
| `groq` | `GROQ_API_KEY` | fast, cheap |
| `deepseek` | `DEEPSEEK_API_KEY` | cheap, good at code |
| `gemini` | `GEMINI_API_KEY` | via Google's OpenAI-compatible endpoint |
| `mistral` | `MISTRAL_API_KEY` | |
| `xai` | `XAI_API_KEY` | |
| `cerebras` | `CEREBRAS_API_KEY` | very fast inference |
| `together` | `TOGETHER_API_KEY` | many open models |
| `fireworks` | `FIREWORKS_API_KEY` | |
| `perplexity` | `PERPLEXITY_API_KEY` | search-grounded |
| `novita` | `NOVITA_API_KEY` | |
| `local` | nothing | Ollama, commented out |
| `lmstudio` | nothing | LM Studio, commented out |

The provider without a key is `default`, so the first request asks for its key
rather than picking a provider for you. Switch with `shellout -p groq ...`.

- Model names change. List the ones your key can reach with
  `curl -s https://api.groq.com/openai/v1/models -H "Authorization: Bearer $GROQ_API_KEY"`
  and put an id in `model`.
- Profiles are named. `default` picks the one used without `-p`; `SHELLOUT_PROFILE`
  and then `-p NAME` override it, in that order.
- `api_key_env` holds the **name** of the environment variable that holds the key,
  never the key itself, so the file is safe to keep in dotfiles. Leave it out for
  local servers that need no authentication. Missing key is reported per profile:
  `shellout: GROQ_API_KEY is not set (profile: groq)`.
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
