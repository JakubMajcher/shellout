// SPDX-License-Identifier: GPL-3.0-or-later

package config

// Profiles marked "verified" answered an HTTP 401/403/400 on an unauthenticated
// POST to <base_url>/chat/completions on 2026-10-05, which means the endpoint
// exists and is reachable. Model names change often; list yours with
//
//	curl -s <base_url>/models -H "Authorization: Bearer $YOUR_KEY_ENV"
//
// and put the id in the model field. Profiles that need no API key are commented
// out, because they are only useful on some machines.
const sampleConfig = `# shellout config. Docs: https://github.com/JakubMajcher/shellout

# Profile used without -p. The SHELLOUT_PROFILE environment variable overrides it.
# Switch profiles with: shellout -p NAME ...
default = "openrouter"

# ---------------------------------------------------------------------------
# OpenRouter: one key, hundreds of models. Swap the model, not the profile.
# Verified. https://openrouter.ai/models
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
# OpenAI. Verified.
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
# Groq: fast and cheap. Verified.
# ---------------------------------------------------------------------------
[profiles.groq]
base_url = "https://api.groq.com/openai/v1"
model = "llama-3.3-70b-versatile"
api_key_env = "GROQ_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# DeepSeek: cheap, good at code. Verified.
# ---------------------------------------------------------------------------
[profiles.deepseek]
base_url = "https://api.deepseek.com/v1"
model = "deepseek-chat"
api_key_env = "DEEPSEEK_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Google Gemini through its OpenAI-compatible endpoint. Verified.
# ---------------------------------------------------------------------------
[profiles.gemini]
base_url = "https://generativelanguage.googleapis.com/v1beta/openai"
model = "gemini-2.5-flash"
api_key_env = "GEMINI_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Mistral. Verified.
# ---------------------------------------------------------------------------
[profiles.mistral]
base_url = "https://api.mistral.ai/v1"
model = "mistral-large-latest"
api_key_env = "MISTRAL_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# xAI Grok. Verified.
# ---------------------------------------------------------------------------
[profiles.xai]
base_url = "https://api.x.ai/v1"
model = "grok-4"
api_key_env = "XAI_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Cerebras: very fast inference. Verified.
# ---------------------------------------------------------------------------
[profiles.cerebras]
base_url = "https://api.cerebras.ai/v1"
model = "llama-3.3-70b"
api_key_env = "CEREBRAS_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Together AI: many open models. Verified.
# ---------------------------------------------------------------------------
[profiles.together]
base_url = "https://api.together.xyz/v1"
model = "meta-llama/Llama-3.3-70B-Instruct-Turbo"
api_key_env = "TOGETHER_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Fireworks AI. Verified.
# ---------------------------------------------------------------------------
[profiles.fireworks]
base_url = "https://api.fireworks.ai/inference/v1"
model = "accounts/fireworks/models/llama-v3p3-70b-instruct"
api_key_env = "FIREWORKS_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Perplexity: search-grounded models. Verified.
# ---------------------------------------------------------------------------
[profiles.perplexity]
base_url = "https://api.perplexity.ai"
model = "sonar"
api_key_env = "PERPLEXITY_API_KEY"
timeout = "60s"

# ---------------------------------------------------------------------------
# Novita AI. Verified.
# ---------------------------------------------------------------------------
[profiles.novita]
base_url = "https://api.novita.ai/v3/openai"
model = "meta-llama/llama-3.3-70b-instruct"
api_key_env = "NOVITA_API_KEY"
timeout = "60s"

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
