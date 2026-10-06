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
// after that, so it gets the same treatment with the key URL for whichever
// profile is selected. Everything else keeps the plain error form from D12.
func configProblem(path, flagProfile string, err error) int {
	var created *config.CreatedError
	if errors.As(err, &created) {
		shell, rc := startupFile()
		fmt.Fprint(os.Stderr, firstRunNotice(created.Path, shell, rc,
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

// startupFile returns the user's shell and its startup file, so the guide can
// hand over a ready to paste line. Either may be empty when the shell is
// something shellout has no integration for.
func startupFile() (shell, rcPath string) {
	info := sysinfo.Collect()
	if info.Shell == "sh" {
		return "", ""
	}
	rc, err := shellinit.RCFile(info.Shell, os.Getenv("HOME"))
	if err != nil {
		return info.Shell, ""
	}
	return info.Shell, rc
}

// firstRunNotice is what the very first run prints. It is not formatted as an
// error: nothing went wrong, shellout just has nothing to talk to yet.
//
// The order is deliberate. Integration comes first because without it cd and
// export silently do nothing, which is the most confusing failure in the tool
// and the hardest to diagnose from a command that ran fine. The key comes
// second because without it there is no suggestion at all.
func firstRunNotice(configPath, shell, rcPath, profile, apiKeyEnv string) string {
	var b strings.Builder

	b.WriteString("shellout needs two things before it can suggest a command.\n\n")

	b.WriteString("  1. let it run commands in this shell, so cd and export keep working\n")
	if rcPath != "" {
		fmt.Fprintf(&b, "     echo 'eval \"$(%s init %s)\"' >> %s\n", app.Name, shell, rcPath)
	} else {
		fmt.Fprintf(&b, "     eval \"$(%s init %s)\"\n", app.Name, shell)
	}
	fmt.Fprintf(&b, "     without it every command runs in a child process and cd is wasted\n\n")

	b.WriteString("  2. add your API key\n")
	fmt.Fprintf(&b, "     export %s=...\n", apiKeyEnv)
	if url, ok := keyURLs[profile]; ok {
		fmt.Fprintf(&b, "     keys for %s: %s\n", profile, url)
	} else {
		fmt.Fprintf(&b, "     see the provider's site for a key; profile: %s\n", profile)
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "  a sample config is at %s\n", configPath)
	fmt.Fprintf(&b, "  change the default profile there, or switch with %s -p NAME ...\n", app.Name)
	return b.String()
}

// missingKeyNotice is printed when a profile is selected but its key is absent,
// which is the second wall a new user hits right after the first run.
func missingKeyNotice(profile, apiKeyEnv string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s is not set (profile: %s)\n", apiKeyEnv, profile)
	if url, ok := keyURLs[profile]; ok {
		fmt.Fprintf(&b, "  set it with:  export %s=...\n", apiKeyEnv)
		fmt.Fprintf(&b, "  keys:         %s\n", url)
	} else {
		fmt.Fprintf(&b, "  set it with:  export %s=...\n", apiKeyEnv)
	}
	fmt.Fprintf(&b, "  other profiles: %s -p NAME ...\n", app.Name)
	return b.String()
}
