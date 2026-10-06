// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"strconv"
	"strings"
	"testing"

	"github.com/JakubMajcher/shellout/internal/app"
	"github.com/JakubMajcher/shellout/internal/config"
)

// The first run is the only place a new user learns what shellout needs. These
// tests pin the order, the content, and the two things it must not do.
func TestFirstRunNoticeOrder(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", "/home/u")
	got := firstRunNotice("/home/u/.config/shellout/config.toml",
		config.SampleDefault, config.SampleKeyEnv)

	iConfig := strings.Index(got, "config.toml")
	iKey := strings.Index(got, "OPENAI_API_KEY")
	iInit := strings.Index(got, "init zsh")

	if iConfig < 0 {
		t.Fatalf("does not say where the config is:\n%s", got)
	}
	if iKey < 0 {
		t.Fatalf("does not ask for the key:\n%s", got)
	}
	if iInit < 0 {
		t.Fatalf("does not offer the integration:\n%s", got)
	}
	if iConfig > iKey {
		t.Error("the config must be reported before the key")
	}
	if iKey > iInit {
		t.Error("the key must come before the integration")
	}
}

// Step 1 must name the current default, because the key asked for in step 2
// depends on it.
func TestFirstRunNoticeNamesTheDefaultProfile(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	got := firstRunNotice("/tmp/config.toml", "openai", "OPENAI_API_KEY")
	if !strings.Contains(got, `"openai"`) {
		t.Errorf("does not name the default profile:\n%s", got)
	}
	if !strings.Contains(got, "-p NAME") {
		t.Errorf("does not offer -p NAME as the alternative:\n%s", got)
	}
	if !strings.Contains(got, "sample config") {
		t.Errorf("does not say a config was written:\n%s", got)
	}
}

// The key instruction has to follow from the default, so a profile other than
// the sample default must change the variable it names.
func TestFirstRunNoticeKeyFollowsTheProfile(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	got := firstRunNotice("/tmp/config.toml", "groq", "GROQ_API_KEY")
	if !strings.Contains(got, "GROQ_API_KEY") {
		t.Errorf("does not name the key for the profile:\n%s", got)
	}
	if strings.Contains(got, "OPENAI_API_KEY") {
		t.Errorf("leaked the sample default key:\n%s", got)
	}
	if !strings.Contains(got, "https://console.groq.com/keys") {
		t.Errorf("no key URL for groq:\n%s", got)
	}
}

// Nothing went wrong on the first run, so it must not read like a crash.
func TestFirstRunNoticeIsNotFormattedAsAnError(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	got := firstRunNotice("/tmp/config.toml", "openai", "OPENAI_API_KEY")
	if strings.HasPrefix(got, app.Name+":") {
		t.Errorf("first run uses the error prefix:\n%s", got)
	}
	for _, bad := range []string{"panic", "traceback", "error:"} {
		if strings.Contains(strings.ToLower(got), bad) {
			t.Errorf("first run output mentions %q:\n%s", bad, got)
		}
	}
}

// The paste line must be relative to the home directory, not an absolute path.
func TestFirstRunNoticeUsesTildeNotHomeDirectory(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	t.Setenv("HOME", "/home/someone")
	got := firstRunNotice("/tmp/config.toml", "openai", "OPENAI_API_KEY")
	if !strings.Contains(got, "' >> ~/.zshrc") {
		t.Errorf("integration line does not use ~:\n%s", got)
	}
	if strings.Contains(got, "/home/someone") {
		t.Errorf("prints the home directory into a pasteable line:\n%s", got)
	}
}

// Integration is the optional step, so it has to be marked as such.
func TestFirstRunNoticeMarksIntegrationOptional(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	got := firstRunNotice("/tmp/config.toml", "openai", "OPENAI_API_KEY")
	if !strings.Contains(strings.ToLower(got), "optional") {
		t.Errorf("integration is not marked optional:\n%s", got)
	}
}

// A shell with no integration must not produce a broken line.
func TestFirstRunNoticeWithoutShellSupport(t *testing.T) {
	t.Setenv("SHELL", "")
	got := firstRunNotice("/tmp/config.toml", "openai", "OPENAI_API_KEY")
	if strings.Contains(got, "init ") {
		t.Errorf("suggests an integration for a shell it does not support:\n%s", got)
	}
	if !strings.Contains(got, "OPENAI_API_KEY") {
		t.Errorf("lost the key step:\n%s", got)
	}
}

// An unknown provider must not get a guessed or wrong link.
func TestFirstRunNoticeUnknownProvider(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	got := firstRunNotice("/tmp/config.toml", "myownllm", "MY_LLM_KEY")
	if strings.Contains(got, "https://") {
		t.Errorf("invented a URL for an unknown provider:\n%s", got)
	}
	if !strings.Contains(got, "MY_LLM_KEY") || !strings.Contains(got, "myownllm") {
		t.Errorf("must still name the variable and profile:\n%s", got)
	}
}

// The profile count in step 1 comes from the sample, so it cannot go stale.
func TestFirstRunNoticeProfileCountMatchesSample(t *testing.T) {
	t.Setenv("SHELL", "/bin/zsh")
	got := firstRunNotice("/tmp/config.toml", "openai", "OPENAI_API_KEY")
	want := config.SampleProfileCount()
	if want < 10 {
		t.Fatalf("sample ships only %d profiles", want)
	}
	if !strings.Contains(got, "with "+strconv.Itoa(want)+" profiles") {
		t.Errorf("does not report %d profiles:\n%s", want, got)
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
	for _, name := range config.SampleProfiles() {
		if _, ok := keyURLs[name]; !ok {
			t.Errorf("no key URL for shipped profile %q", name)
		}
	}
}
