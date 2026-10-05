// SPDX-License-Identifier: GPL-3.0-or-later

package config

const sampleConfig = `# shellout config. Docs: https://github.com/JakubMajcher/shellout

# Profile used without -p. The SHELLOUT_PROFILE environment variable overrides it.
default = "openai"

[profiles.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-5-mini"
# Name of the environment variable that holds the API key. Never put the key here.
api_key_env = "OPENAI_API_KEY"
timeout = "60s"

# Everything in params is copied as-is into the request body, and shellout sends
# no other optional parameter. Use it for model settings and to cap cost.
# Parameter names differ between providers; use the ones yours accepts.
# OpenRouter: add  provider = { require_parameters = true }  so requests only go
# to providers that enforce the JSON schema.
[profiles.openai.params]
reasoning_effort = "minimal"
max_completion_tokens = 1000

# A local model through Ollama, no API key needed. Use it with: shellout -p local ...
# [profiles.local]
# base_url = "http://localhost:11434/v1"
# model = "qwen2.5-coder:7b"
`
