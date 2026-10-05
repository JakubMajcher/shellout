// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const validConfig = `
default = "openai"

[profiles.openai]
base_url = "https://api.openai.com/v1"
model = "gpt-5-mini"
api_key_env = "OPENAI_API_KEY"

[profiles.openai.params]
reasoning_effort = "minimal"
max_completion_tokens = 1000

[profiles.openai.params.provider]
require_parameters = true

[profiles.local]
base_url = "http://localhost:11434/v1"
model = "qwen2.5-coder:7b"
timeout = "5s"
`

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadDefaultProfile(t *testing.T) {
	p := writeConfig(t, validConfig)
	got, err := Load(p, "", env(map[string]string{"OPENAI_API_KEY": "sk-test"}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "openai" || got.Model != "gpt-5-mini" || got.APIKey != "sk-test" {
		t.Fatalf("unexpected profile: %+v", got)
	}
	if got.Timeout != 60*time.Second {
		t.Fatalf("default timeout = %v, want 60s", got.Timeout)
	}
	if got.Params["reasoning_effort"] != "minimal" {
		t.Fatalf("reasoning_effort = %v", got.Params["reasoning_effort"])
	}
	if got.Params["max_completion_tokens"] != int64(1000) {
		t.Fatalf("max_completion_tokens = %#v", got.Params["max_completion_tokens"])
	}
	provider, ok := got.Params["provider"].(map[string]any)
	if !ok || provider["require_parameters"] != true {
		t.Fatalf("nested params lost: %#v", got.Params["provider"])
	}
}

func TestProfileSelectionOrder(t *testing.T) {
	p := writeConfig(t, validConfig)
	e := env(map[string]string{"OPENAI_API_KEY": "k", "SHELLOUT_PROFILE": "local"})

	got, err := Load(p, "", e)
	if err != nil || got.Name != "local" {
		t.Fatalf("env should override default: %+v, %v", got, err)
	}
	got, err = Load(p, "openai", e)
	if err != nil || got.Name != "openai" {
		t.Fatalf("flag should override env: %+v, %v", got, err)
	}
}

func TestProfileWithoutKeyAndCustomTimeout(t *testing.T) {
	p := writeConfig(t, validConfig)
	got, err := Load(p, "local", env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if got.APIKey != "" || got.Timeout != 5*time.Second {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

func TestUnknownProfile(t *testing.T) {
	p := writeConfig(t, validConfig)
	_, err := Load(p, "nope", env(nil))
	want := `unknown profile "nope" (have: local, openai)`
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}

func TestUnknownKeyIsRejected(t *testing.T) {
	body := strings.Replace(validConfig, "api_key_env", "api_key_evn", 1)
	p := writeConfig(t, body)
	_, err := Load(p, "", env(nil))
	if err == nil || !strings.Contains(err.Error(), `unknown key "profiles.openai.api_key_evn"`) {
		t.Fatalf("err = %v", err)
	}
}

// params is a verbatim passthrough, so anything below it is allowed. A typo in
// the params key itself is still an error, because then the passthrough the user
// intended would be silently ignored.
func TestParamsIsFreeFormButItsNameIsChecked(t *testing.T) {
	deep := `
default = "local"

[profiles.local]
base_url = "http://localhost:11434/v1"
model = "m"
[profiles.local.params.some_future_provider]
anything_goes_here = true
`
	if _, err := Load(writeConfig(t, deep), "", env(nil)); err != nil {
		t.Fatalf("free-form params rejected: %v", err)
	}
	typo := strings.Replace(deep, "params.some_future_provider", "paramss.some_future_provider", 1)
	_, err := Load(writeConfig(t, typo), "", env(nil))
	if err == nil || !strings.Contains(err.Error(), `unknown key "profiles.local.paramss.`) {
		t.Fatalf("params typo not caught: %v", err)
	}
}

func TestForbiddenParams(t *testing.T) {
	for _, key := range []string{"model", "messages", "response_format", "stream"} {
		body := strings.Replace(validConfig, `reasoning_effort = "minimal"`,
			`reasoning_effort = "minimal"`+"\n"+key+` = "x"`, 1)
		p := writeConfig(t, body)
		_, err := Load(p, "", env(map[string]string{"OPENAI_API_KEY": "k"}))
		if err == nil || !strings.Contains(err.Error(), `params may not set "`+key+`"`) {
			t.Fatalf("%s: err = %v", key, err)
		}
	}
}

func TestMissingRequiredField(t *testing.T) {
	body := strings.Replace(validConfig, `model = "gpt-5-mini"`, "", 1)
	p := writeConfig(t, body)
	_, err := Load(p, "", env(map[string]string{"OPENAI_API_KEY": "k"}))
	if err == nil || !strings.Contains(err.Error(), `profile "openai": missing model`) {
		t.Fatalf("err = %v", err)
	}
}

func TestInvalidTimeout(t *testing.T) {
	body := strings.Replace(validConfig, `timeout = "5s"`, `timeout = "soon"`, 1)
	p := writeConfig(t, body)
	_, err := Load(p, "local", env(nil))
	if err == nil || !strings.Contains(err.Error(), `invalid timeout "soon"`) {
		t.Fatalf("err = %v", err)
	}
}

func TestMissingAPIKey(t *testing.T) {
	p := writeConfig(t, validConfig)
	_, err := Load(p, "", env(nil))
	want := "OPENAI_API_KEY is not set (profile: openai)"
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
}

func TestMissingFileWritesLoadableSample(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "config.toml")
	_, err := Load(p, "", env(nil))
	var created *CreatedError
	if !errors.As(err, &created) || created.Path != p {
		t.Fatalf("err = %v, want CreatedError for %s", err, p)
	}
	if err.Error() != "created sample config at "+p+", edit it and run again" {
		t.Fatalf("message = %q", err.Error())
	}
	// The written sample must load with only the default profile's key present.
	got, err := Load(p, "", env(map[string]string{"OPENROUTER_API_KEY": "k"}))
	if err != nil {
		t.Fatalf("sample does not load: %v", err)
	}
	if got.Name != "openrouter" || got.Model == "" || got.Params["provider"] == nil {
		t.Fatalf("unexpected sample profile: %+v", got)
	}

	// The file must stay safe to keep in dotfiles: no key value, and only the
	// variable name.
	body, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	// api_key_env holds the variable name only, never a key value.
	if !strings.Contains(string(body), "api_key_env") {
		t.Fatal("sample must name the key variable instead of holding the key")
	}
	// The OpenRouter note from section D7 has to be in the sample.
	if !strings.Contains(string(body), "require_parameters") {
		t.Fatal("sample must document provider.require_parameters for OpenRouter")
	}
	if info, err := os.Stat(p); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("sample mode = %v, want 0600", info.Mode().Perm())
	}
}

func TestPath(t *testing.T) {
	got, _ := Path(env(map[string]string{"SHELLOUT_CONFIG": "/x/c.toml", "XDG_CONFIG_HOME": "/xdg", "HOME": "/h"}))
	if got != "/x/c.toml" {
		t.Fatalf("SHELLOUT_CONFIG: %s", got)
	}
	got, _ = Path(env(map[string]string{"XDG_CONFIG_HOME": "/xdg", "HOME": "/h"}))
	if got != "/xdg/shellout/config.toml" {
		t.Fatalf("XDG: %s", got)
	}
	got, _ = Path(env(map[string]string{"HOME": "/h"}))
	if got != "/h/.config/shellout/config.toml" {
		t.Fatalf("HOME: %s", got)
	}
	if _, err := Path(env(nil)); err == nil {
		t.Fatal("expected error without HOME")
	}
}
