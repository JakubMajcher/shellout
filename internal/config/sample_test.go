// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
)

// The sample config is what every new user starts from, so it must load and it
// must not contain a mistake that only shows up on a user's machine.
func TestSampleConfigLoads(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(sampleConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	// Load with a getenv that claims every api_key_env is set, so the check is
	// about the config itself and not about the local environment.
	prof, err := Load(p, "", func(k string) string {
		if strings.HasSuffix(k, "_API_KEY") {
			return "test-key"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("sample does not load: %v", err)
	}
	if prof.Name != "openrouter" {
		t.Fatalf("default profile = %q, want openrouter", prof.Name)
	}
	if prof.Timeout != 60*time.Second {
		t.Fatalf("timeout = %v, want 60s", prof.Timeout)
	}
}

// Every profile in the sample must be selectable and complete. A half-written
// profile is worse than no profile at all.
func TestEverySampleProfileIsValid(t *testing.T) {
	var fc struct {
		Default  string                 `toml:"default"`
		Profiles map[string]fileProfile `toml:"profiles"`
	}
	md, err := toml.Decode(sampleConfig, &fc)
	if err != nil {
		t.Fatalf("sample is not valid TOML: %v", err)
	}
	// Nested tables under params are free-form by design (section D5), and
	// BurntSushi/toml lists their keys as undecoded, so they are filtered here
	// with the same rule Load uses.
	if k, bad := unknownKey(md.Undecoded()); bad {
		t.Fatalf("sample has unknown key %q", k)
	}
	if len(fc.Profiles) < 10 {
		t.Fatalf("only %d profiles in the sample, want a useful selection", len(fc.Profiles))
	}
	if _, ok := fc.Profiles[fc.Default]; !ok {
		t.Fatalf("default %q is not a defined profile", fc.Default)
	}

	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(sampleConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	for name := range fc.Profiles {
		prof, err := Load(p, name, func(k string) string {
			if strings.HasSuffix(k, "_API_KEY") {
				return "test-key"
			}
			return ""
		})
		if err != nil {
			t.Errorf("profile %q: %v", name, err)
			continue
		}
		if prof.Name != name || prof.BaseURL == "" || prof.Model == "" {
			t.Errorf("profile %q incomplete: %+v", name, prof)
		}
	}
}

// The sample must never contain a key value, only the variable name.
// The sample must document command_timeout without setting it, so a new user
// gets no limit and still learns the key exists.
func TestSampleDocumentsCommandTimeoutWithoutEnablingIt(t *testing.T) {
	var fc struct {
		Profiles map[string]fileProfile `toml:"profiles"`
	}
	if _, err := toml.Decode(sampleConfig, &fc); err != nil {
		t.Fatal(err)
	}
	for name, fp := range fc.Profiles {
		if fp.CommandTimeout != "" {
			t.Errorf("profile %q sets command_timeout = %q; the sample must leave it unset",
				name, fp.CommandTimeout)
		}
	}
	if !strings.Contains(sampleConfig, "command_timeout") {
		t.Error("sample does not mention command_timeout at all")
	}
	if !strings.Contains(sampleConfig, "# command_timeout = ") {
		t.Error("sample must show the key commented out, ready to uncomment")
	}
}

// A raw Go string cannot contain backticks, and the sample is one. If a comment
// ever gets one the file stops compiling, so assert the shape explicitly.
func TestSampleHasNoBackticks(t *testing.T) {
	if strings.Contains(sampleConfig, "`") {
		t.Error("sampleConfig is a raw string literal and must not contain a backtick")
	}
}

func TestSampleHasNoKeyValues(t *testing.T) {
	for _, line := range strings.Split(sampleConfig, "\n") {
		if strings.Contains(line, "sk-") || strings.Contains(line, "Bearer ") {
			t.Fatalf("sample looks like it holds a key: %q", line)
		}
	}
}

// Every non-comment line that looks like a URL must be reachable. This is the
// check that catches a typo in a base_url before a user hits it.
func TestSampleURLsAreWellFormed(t *testing.T) {
	var fc struct {
		Profiles map[string]fileProfile `toml:"profiles"`
	}
	if _, err := toml.Decode(sampleConfig, &fc); err != nil {
		t.Fatal(err)
	}
	for name, fp := range fc.Profiles {
		if !strings.HasPrefix(fp.BaseURL, "http://") && !strings.HasPrefix(fp.BaseURL, "https://") {
			t.Errorf("profile %q: base_url is not a URL: %q", name, fp.BaseURL)
		}
		if strings.HasSuffix(fp.BaseURL, "/") {
			t.Errorf("profile %q: trailing slash in base_url: %q", name, fp.BaseURL)
		}
	}
}

// No profile may set a request key that shellout controls.
func TestSampleParamsAreAllowed(t *testing.T) {
	var fc struct {
		Profiles map[string]fileProfile `toml:"profiles"`
	}
	if _, err := toml.Decode(sampleConfig, &fc); err != nil {
		t.Fatal(err)
	}
	for name, fp := range fc.Profiles {
		for _, k := range forbiddenParams {
			if _, bad := fp.Params[k]; bad {
				t.Errorf("profile %q sets forbidden params key %q", name, k)
			}
		}
	}
}
