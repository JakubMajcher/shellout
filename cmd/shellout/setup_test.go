// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"strings"
	"testing"

	"github.com/JakubMajcher/shellout/internal/app"
)

// The first run is the only place a new user learns what shellout needs. These
// tests pin the two things it must do and the two it must not.
func TestFirstRunNoticeOrderAndContent(t *testing.T) {
	got := firstRunNotice("/home/u/.config/shellout/config.toml", "zsh",
		"/home/u/.zshrc", "openai", "OPENAI_API_KEY")

	// Integration is step 1, the key is step 2.
	iInit := strings.Index(got, "shellout init zsh")
	iKey := strings.Index(got, "export OPENAI_API_KEY")
	if iInit < 0 {
		t.Fatalf("no init suggestion:\n%s", got)
	}
	if iKey < 0 {
		t.Fatalf("no key instruction:\n%s", got)
	}
	if iInit > iKey {
		t.Error("integration must come before the key")
	}
	// The paste line must be complete and quoted, not a bare hint.
	if !strings.Contains(got, `echo 'eval "$(shellout init zsh)"' >> /home/u/.zshrc`) {
		t.Errorf("not a pasteable line:\n%s", got)
	}
	if !strings.Contains(got, "https://platform.openai.com/api-keys") {
		t.Errorf("no key URL for a known provider:\n%s", got)
	}
	if !strings.Contains(got, "/home/u/.config/shellout/config.toml") {
		t.Error("does not say where the config went")
	}
}

// Nothing went wrong on the first run, so it must not read like a crash.
func TestFirstRunNoticeIsNotFormattedAsAnError(t *testing.T) {
	got := firstRunNotice("/tmp/config.toml", "zsh", "/home/u/.zshrc", "openai", "OPENAI_API_KEY")
	if strings.HasPrefix(got, app.Name+":") {
		t.Errorf("first run uses the error prefix:\n%s", got)
	}
	for _, bad := range []string{"panic", "traceback", "error:"} {
		if strings.Contains(strings.ToLower(got), bad) {
			t.Errorf("first run output mentions %q:\n%s", bad, got)
		}
	}
}

// An unknown shell has no startup file to name, but the command still works.
func TestFirstRunNoticeWithoutStartupFile(t *testing.T) {
	got := firstRunNotice("/tmp/config.toml", "sh", "", "openai", "OPENAI_API_KEY")
	if strings.Contains(got, ">>") {
		t.Errorf("must not suggest appending to a file it does not know:\n%s", got)
	}
	if !strings.Contains(got, `eval "$(shellout init sh)"`) {
		t.Errorf("should still show the eval line:\n%s", got)
	}
}

// A provider shellout does not know must not get a guessed or wrong link.
func TestFirstRunNoticeUnknownProvider(t *testing.T) {
	got := firstRunNotice("/tmp/config.toml", "zsh", "/home/u/.zshrc", "myownllm", "MY_LLM_KEY")
	if strings.Contains(got, "https://") {
		t.Errorf("invented a URL for an unknown provider:\n%s", got)
	}
	if !strings.Contains(got, "MY_LLM_KEY") || !strings.Contains(got, "myownllm") {
		t.Errorf("must still name the variable and profile:\n%s", got)
	}
}

// The missing key notice names the selected profile, not the sample default.
func TestMissingKeyNoticeNamesTheSelectedProfile(t *testing.T) {
	got := missingKeyNotice("groq", "GROQ_API_KEY")
	if !strings.Contains(got, "GROQ_API_KEY is not set (profile: groq)") {
		t.Errorf("lost the original wording:\n%s", got)
	}
	if !strings.Contains(got, "export GROQ_API_KEY=...") {
		t.Errorf("does not say how to set it:\n%s", got)
	}
	if !strings.Contains(got, "https://console.groq.com/keys") {
		t.Errorf("no key URL for groq:\n%s", got)
	}
	if strings.Contains(got, "OPENAI") {
		t.Error("leaked the sample default into another profile's notice")
	}
}

func TestMissingKeyNoticeUnknownProvider(t *testing.T) {
	got := missingKeyNotice("myownllm", "MY_LLM_KEY")
	if strings.Contains(got, "https://") {
		t.Errorf("invented a URL for an unknown provider:\n%s", got)
	}
	if !strings.Contains(got, "MY_LLM_KEY is not set (profile: myownllm)") {
		t.Errorf("lost the wording:\n%s", got)
	}
}

// Every profile the sample ships must have a key URL, or a new user choosing it
// gets a worse message than the default one.
func TestEverySampleProfileHasAKeyURL(t *testing.T) {
	profiles := map[string]string{
		"openai": "OPENAI_API_KEY", "openrouter": "OPENROUTER_API_KEY",
		"groq": "GROQ_API_KEY", "deepseek": "DEEPSEEK_API_KEY",
		"gemini": "GEMINI_API_KEY", "mistral": "MISTRAL_API_KEY",
		"xai": "XAI_API_KEY", "cerebras": "CEREBRAS_API_KEY",
		"together": "TOGETHER_API_KEY", "fireworks": "FIREWORKS_API_KEY",
		"perplexity": "PERPLEXITY_API_KEY", "novita": "NOVITA_API_KEY",
	}
	for name, env := range profiles {
		if _, ok := keyURLs[name]; !ok {
			t.Errorf("no key URL for shipped profile %q (%s)", name, env)
		}
	}
}
