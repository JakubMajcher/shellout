// SPDX-License-Identifier: GPL-3.0-or-later

package config

// Model names change often; list the ones your key can reach with
//
//	curl -s <base_url>/models -H "Authorization: Bearer $YOUR_KEY_ENV"
//
// and put an id in the model field. Profiles that need no API key are commented
// out, because they are only useful on some machines.
const sampleConfig = `# shellout config. Docs: https://github.com/JakubMajcher/shellout

# Profile used without -p. The SHELLOUT_PROFILE environment variable overrides it.
# Switch profiles with: shellout -p NAME ...
default = "openrouter"

# ---------------------------------------------------------------------------
# OpenRouter: one key, hundreds of models. Swap the model, not the profile.
# https://openrouter.ai/models
# ---------------------------------------------------------------------------
[profiles.openrouter]
base_url = "https://openrouter.ai/api/v1"
model = "openai/gpt-5.4-mini"
api_key_env = "OPENROUTER_API_KEY"
timeout = "60s"

[profiles.openrouter.params]
max_completion_tokens = 1000
# Required. Without it OpenRouter may route the request to a provider that
# silently ignores the JSON schema, and shellout then gets prose, not a command.
provider = { require_parameters = true }

# ---------------------------------------------------------------------------
# OpenAI
# ---------------------------------------------------------------------------
[profiles.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-5-mini"
# Name of the environment variable that holds the API key. Never put the key here.
api_key_env = "OPENAI_API_KEY"
timeout = "60s"

[profiles.openai.params]
reasoning_effort = "minimal"
max_completion_tokens = 1000

# ---------------------------------------------------------------------------
# Groq: fast and cheap
# ---------------------------------------------------------------------------
[profiles.groq]
base_url = "https://api.groq.com/openai/v1"
model = "llama-3.3-70b-versatile"
api_key_env = "GROQ_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# DeepSeek: cheap, good at code
# ---------------------------------------------------------------------------
[profiles.deepseek]
base_url = "https://api.deepseek.com/v1"
model = "deepseek-chat"
api_key_env = "DEEPSEEK_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Google Gemini through its OpenAI-compatible endpoint
# ---------------------------------------------------------------------------
[profiles.gemini]
base_url = "https://generativelanguage.googleapis.com/v1beta/openai"
model = "gemini-2.5-flash"
api_key_env = "GEMINI_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Mistral
# ---------------------------------------------------------------------------
[profiles.mistral]
base_url = "https://api.mistral.ai/v1"
model = "mistral-large-latest"
api_key_env = "MISTRAL_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# xAI Grok
# ---------------------------------------------------------------------------
[profiles.xai]
base_url = "https://api.x.ai/v1"
model = "grok-4"
api_key_env = "XAI_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Cerebras: very fast inference
# ---------------------------------------------------------------------------
[profiles.cerebras]
base_url = "https://api.cerebras.ai/v1"
model = "llama-3.3-70b"
api_key_env = "CEREBRAS_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Together AI: many open models
# ---------------------------------------------------------------------------
[profiles.together]
base_url = "https://api.together.xyz/v1"
model = "meta-llama/Llama-3.3-70B-Instruct-Turbo"
api_key_env = "TOGETHER_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Fireworks AI
# ---------------------------------------------------------------------------
[profiles.fireworks]
base_url = "https://api.fireworks.ai/inference/v1"
model = "accounts/fireworks/models/llama-v3p3-70b-instruct"
api_key_env = "FIREWORKS_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Perplexity: search-grounded models
# ---------------------------------------------------------------------------
[profiles.perplexity]
base_url = "https://api.perplexity.ai"
model = "sonar"
api_key_env = "PERPLEXITY_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Novita AI
# ---------------------------------------------------------------------------
[profiles.novita]
base_url = "https://api.novita.ai/v3/openai"
model = "meta-llama/llama-3.3-70b-instruct"
api_key_env = "NOVITA_API_KEY"
timeout = "60s"

# ===========================================================================
# Optional per profile
# ===========================================================================

# The "timeout" above caps how long the model may think. "command_timeout" caps
# how long the approved command may run, and is unset by default: a cap that
# fires half way through can leave a truncated archive or a partial transfer
# behind, and you can see whether a command is long. Uncomment to add one.
#
# When it fires, the whole command is killed and shellout exits 124, the same
# code the "timeout" command uses.
# command_timeout = "10m"

# ===========================================================================
# No API key needed. Uncomment the one you actually run, then use -p.
# ===========================================================================

# Ollama. No key. Use it with: shellout -p local ...
# [profiles.local]
# base_url = "http://localhost:11434/v1"
# model = "qwen2.5-coder:7b"

# LM Studio. No key. Its server listens on 1234 by default.
# [profiles.lmstudio]
# base_url = "http://localhost:1234/v1"
# model = "local-model"
`
