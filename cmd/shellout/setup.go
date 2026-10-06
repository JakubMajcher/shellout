// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/JakubMajcher/shellout/internal/app"
	"github.com/JakubMajcher/shellout/internal/config"
	"github.com/JakubMajcher/shellout/internal/shellinit"
	"github.com/JakubMajcher/shellout/internal/sysinfo"
)

// keyURLs points at where the key for a known provider is issued. Used only to
// turn "X is not set" into something actionable. An unknown provider falls back
// to a generic hint rather than a wrong link.
var keyURLs = map[string]string{
	"openai":     "https://platform.openai.com/api-keys",
	"openrouter": "https://openrouter.ai/settings/keys",
	"groq":       "https://console.groq.com/keys",
	"deepseek":   "https://platform.deepseek.com/api_keys",
	"gemini":     "https://aistudio.google.com/apikey",
	"mistral":    "https://console.mistral.ai/api-keys",
	"xai":        "https://console.x.ai",
	"cerebras":   "https://cloud.cerebras.ai",
	"together":   "https://api.together.ai/settings/api-keys",
	"fireworks":  "https://fireworks.ai/api-keys",
	"perplexity": "https://www.perplexity.ai/settings/api",
	"novita":     "https://novita.ai/dashboard/key-manager",
}

// configProblem turns a config failure into something a person can act on.
//
// The first run is not a failure at all: the file was just written. It gets an
// ordered setup guide with no error prefix. A missing key is the wall right
// after that, so it gets the same treatment. Everything else keeps the plain
// error form from D12.
func configProblem(flagProfile string, err error) int {
	var created *config.CreatedError
	if errors.As(err, &created) {
		fmt.Fprint(os.Stderr, firstRunNotice(created.Path,
			config.SampleDefault, config.SampleKeyEnv))
		return 2
	}
	var missing *config.MissingKeyError
	if errors.As(err, &missing) {
		fmt.Fprint(os.Stderr, missingKeyNotice(missing.Profile, missing.Env))
		return 2
	}
	return fail(2, err.Error())
}

// firstRunNotice is what the very first run prints. It is not formatted as an
// error: nothing went wrong, shellout just has nothing to talk to yet.
//
// The order follows what the user has to act on. The config exists already, so
// that is reported first, with the default it currently uses, because everything
// after it depends on that choice. The key is second because without it there is
// no suggestion at all. Integration is last and marked optional, because it is
// the one step that is not required to get started.
func firstRunNotice(configPath, profile, apiKeyEnv string) string {
	shell, rcPath := startupFile()
	var b strings.Builder

	fmt.Fprintf(&b, "%s wrote a sample config with %d profiles, at:\n", app.Name, config.SampleProfileCount())
	fmt.Fprintf(&b, "    %s\n\n", configPath)
	fmt.Fprintf(&b, "It currently defaults to %q. Change it there, or pick another per call:\n", profile)
	fmt.Fprintf(&b, "    %s -p NAME ...\n\n", app.Name)

	fmt.Fprintf(&b, "Now add the API key for that profile:\n")
	fmt.Fprintf(&b, "    export %s=...\n", apiKeyEnv)
	if url, ok := keyURLs[profile]; ok {
		fmt.Fprintf(&b, "    keys: %s\n\n", url)
	} else {
		fmt.Fprintf(&b, "    get a key from your provider\n\n")
	}

	if shell != "" {
		fmt.Fprintf(&b, "Optional, and worth it: let commands run in this shell so that cd\n")
		fmt.Fprintf(&b, "and export keep working, and history shows the real command.\n")
		if rcPath != "" {
			fmt.Fprintf(&b, "    echo 'eval \"$(%s init %s)\"' >> %s\n", app.Name, shell, rcPath)
		} else {
			fmt.Fprintf(&b, "    eval \"$(%s init %s)\"\n", app.Name, shell)
		}
	}
	return b.String()
}

// startupFile returns the user's shell and its startup file. Both may be empty:
// /bin/sh is what Collect reports when $SHELL is unset, and there is no
// integration for it.
func startupFile() (shell, rcPath string) {
	info := sysinfo.Collect()
	if info.Shell == "sh" {
		return "", ""
	}
	rc, err := shellinit.RCPath(info.Shell)
	if err != nil {
		return info.Shell, ""
	}
	return info.Shell, rc
}

// missingKeyNotice is printed when a profile is selected but its key is absent,
// which is the wall right after the first run.
func missingKeyNotice(profile, apiKeyEnv string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is not set (profile: %s)\n", apiKeyEnv, profile)
	fmt.Fprintf(&b, "  set it with:  export %s=...\n", apiKeyEnv)
	if url, ok := keyURLs[profile]; ok {
		fmt.Fprintf(&b, "  keys:         %s\n", url)
	}
	fmt.Fprintf(&b, "  other profiles: %s -p NAME ...\n", app.Name)
	return b.String()
}
