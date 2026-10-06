// SPDX-License-Identifier: GPL-3.0-or-later

// Package config loads the TOML config file and selects a profile.
package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/JakubMajcher/shellout/internal/app"
)

const defaultTimeout = 60 * time.Second

// forbiddenParams are request keys that shellout sets itself.
var forbiddenParams = []string{"model", "messages", "response_format", "stream"}

// Profile is one resolved, validated profile.
type Profile struct {
	Name      string
	BaseURL   string
	Model     string
	APIKeyEnv string
	APIKey    string // value read from the APIKeyEnv variable; empty when APIKeyEnv is empty
	Timeout   time.Duration
	// CommandTimeout caps how long the approved command may run. Zero, the
	// default, means no cap. It is deliberately a separate setting from
	// Timeout: sharing one key would let a 60s API timeout silently start
	// killing commands after a minute.
	CommandTimeout time.Duration
	Params         map[string]any
}

// CreatedError means the config file did not exist and a sample was written.
type CreatedError struct{ Path string }

func (e *CreatedError) Error() string {
	return fmt.Sprintf("created sample config at %s, edit it and run again", e.Path)
}

type fileConfig struct {
	Default  string                 `toml:"default"`
	Profiles map[string]fileProfile `toml:"profiles"`
}

type fileProfile struct {
	BaseURL        string         `toml:"base_url"`
	Model          string         `toml:"model"`
	APIKeyEnv      string         `toml:"api_key_env"`
	Timeout        string         `toml:"timeout"`
	CommandTimeout string         `toml:"command_timeout"`
	Params         map[string]any `toml:"params"`
}

// Path returns the config file location.
func Path(getenv func(string) string) (string, error) {
	if p := getenv("SHELLOUT_CONFIG"); p != "" {
		return p, nil
	}
	if x := getenv("XDG_CONFIG_HOME"); x != "" {
		return filepath.Join(x, app.Name, "config.toml"), nil
	}
	home := getenv("HOME")
	if home == "" {
		return "", errors.New("cannot locate config: HOME is not set")
	}
	return filepath.Join(home, ".config", app.Name, "config.toml"), nil
}

// Load reads the config at path and returns the selected profile.
// Profile choice: flagProfile, then $SHELLOUT_PROFILE, then "default".
func Load(path, flagProfile string, getenv func(string) string) (Profile, error) {
	var fc fileConfig
	md, err := toml.DecodeFile(path, &fc)
	if errors.Is(err, fs.ErrNotExist) {
		if werr := writeSample(path); werr != nil {
			return Profile{}, fmt.Errorf("config %s: %w", path, werr)
		}
		return Profile{}, &CreatedError{Path: path}
	}
	if err != nil {
		return Profile{}, fmt.Errorf("config %s: %w", path, err)
	}
	if k, bad := unknownKey(md.Undecoded()); bad {
		return Profile{}, fmt.Errorf("config %s: unknown key %q", path, k)
	}

	name := flagProfile
	if name == "" {
		name = getenv("SHELLOUT_PROFILE")
	}
	if name == "" {
		name = fc.Default
	}
	if name == "" {
		return Profile{}, fmt.Errorf("config %s: no profile selected; set \"default\" or use -p", path)
	}
	fp, ok := fc.Profiles[name]
	if !ok {
		return Profile{}, fmt.Errorf("unknown profile %q (have: %s)", name, strings.Join(profileNames(fc.Profiles), ", "))
	}

	if fp.BaseURL == "" {
		return Profile{}, fmt.Errorf("config %s: profile %q: missing base_url", path, name)
	}
	if fp.Model == "" {
		return Profile{}, fmt.Errorf("config %s: profile %q: missing model", path, name)
	}
	for _, k := range forbiddenParams {
		if _, bad := fp.Params[k]; bad {
			return Profile{}, fmt.Errorf("config %s: profile %q: params may not set %q", path, name, k)
		}
	}
	timeout := defaultTimeout
	if fp.Timeout != "" {
		d, err := time.ParseDuration(fp.Timeout)
		if err != nil || d <= 0 {
			return Profile{}, fmt.Errorf("config %s: profile %q: invalid timeout %q", path, name, fp.Timeout)
		}
		timeout = d
	}

	// Absent or "0" means no limit. A negative value is rejected rather than
	// treated as no limit, because -1 reads like a mistake.
	commandTimeout := time.Duration(0)
	if fp.CommandTimeout != "" {
		d, err := time.ParseDuration(fp.CommandTimeout)
		if err != nil || d < 0 {
			return Profile{}, fmt.Errorf("config %s: profile %q: invalid command_timeout %q", path, name, fp.CommandTimeout)
		}
		commandTimeout = d
	}

	p := Profile{
		Name:           name,
		BaseURL:        fp.BaseURL,
		Model:          fp.Model,
		APIKeyEnv:      fp.APIKeyEnv,
		Timeout:        timeout,
		CommandTimeout: commandTimeout,
		Params:         fp.Params,
	}
	if p.APIKeyEnv != "" {
		p.APIKey = getenv(p.APIKeyEnv)
		if p.APIKey == "" {
			return Profile{}, fmt.Errorf("%s is not set (profile: %s)", p.APIKeyEnv, name)
		}
	}
	return p, nil
}

func profileNames(m map[string]fileProfile) []string {
	names := make([]string, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// unknownKey returns the first key that is neither a known top-level key nor a
// known profile key.
//
// Keys below a params table are skipped: params is a verbatim passthrough, so
// everything under it is by definition allowed. BurntSushi/toml decodes nested
// tables into map[string]any but does not mark the keys inside them as decoded,
// so md.Undecoded() lists them even though they were read. Rejecting them would
// break every profile with a nested params table, such as the OpenRouter
// provider.require_parameters setting in the sample config.
func unknownKey(keys []toml.Key) (string, bool) {
	for _, k := range keys {
		parts := strings.Split(k.String(), ".")
		// profiles.<name>.params.<anything...> is free-form.
		if len(parts) > 3 && parts[0] == "profiles" && parts[2] == "params" {
			continue
		}
		return k.String(), true
	}
	return "", false
}

func writeSample(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(sampleConfig), 0o600)
}
