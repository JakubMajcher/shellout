// SPDX-License-Identifier: GPL-3.0-or-later

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The README tells people that -p changes nothing on disk and that the config is
// the only place a default is stored. That promise is only worth something if it
// holds, so it is checked here rather than left to a comment.
func TestSelectingAProfileNeverWritesToTheConfig(t *testing.T) {
	p := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(p, []byte(validConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}

	env := func(k string) string {
		if strings.HasSuffix(k, "_API_KEY") {
			return "k"
		}
		return ""
	}

	// Every way of choosing a profile, including the one that does not exist in
	// the file and the one that fails validation.
	for _, name := range []string{"local", "openai", "nonexistent"} {
		if _, err := Load(p, name, env); err != nil {
			continue
		}
	}

	after, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("the config file changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// The README documents a three step precedence: -p, then SHELLOUT_PROFILE, then
// default. Each step must actually beat the one below it.
func TestDocumentedPrecedence(t *testing.T) {
	p := writeConfig(t, validConfig)

	// withKey satisfies every api_key_env and may carry a profile override, so
	// these cases differ only in which name gets picked.
	withKey := func(profile string) func(string) string {
		return func(k string) string {
			if k == "SHELLOUT_PROFILE" {
				return profile
			}
			if strings.HasSuffix(k, "_API_KEY") {
				return "k"
			}
			return ""
		}
	}

	// default alone
	got, err := Load(p, "", withKey(""))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "openai" {
		t.Errorf("no flag, no env: got %q, want openai from default", got.Name)
	}

	// env beats default
	got, err = Load(p, "", withKey("local"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "local" {
		t.Errorf("env should beat default: got %q, want local", got.Name)
	}

	// flag beats env
	got, err = Load(p, "openai", withKey("local"))
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "openai" {
		t.Errorf("flag should beat env: got %q, want openai", got.Name)
	}
}

// The README says an unset api_key_env is what the local profiles rely on, so a
// profile without one must load with no key in the environment at all.
func TestProfileWithoutKeyEnvNeedsNoKey(t *testing.T) {
	p := writeConfig(t, validConfig)
	got, err := Load(p, "local", env(nil))
	if err != nil {
		t.Fatalf("a profile with no api_key_env must not need a key: %v", err)
	}
	if got.APIKey != "" || got.APIKeyEnv != "" {
		t.Errorf("unexpected key handling: %+v", got)
	}
}
