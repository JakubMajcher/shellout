// SPDX-License-Identifier: GPL-3.0-or-later

package llm

import (
	"strings"
	"testing"

	"github.com/JakubMajcher/shellout/internal/sysinfo"
)

var testInfo = sysinfo.Info{OS: "linux", Arch: "amd64", Distro: "Arch Linux", Shell: "zsh", ShellPath: "/bin/zsh", Flavor: "GNU"}

func TestSystemPromptEnvironmentLine(t *testing.T) {
	p := SystemPrompt(testInfo, false)
	want := "Environment: OS linux Arch Linux, arch amd64, shell zsh, core utilities GNU."
	if !strings.Contains(p, want) {
		t.Fatalf("prompt missing %q:\n%s", want, p)
	}
	if !strings.Contains(p, `start the description with "Warning:"`) {
		t.Fatal("prompt missing the Warning rule")
	}
}

func TestSystemPromptStandaloneParagraph(t *testing.T) {
	const marker = "The command runs in a child process."
	if !strings.Contains(SystemPrompt(testInfo, false), marker) {
		t.Fatal("standalone prompt must explain the child process")
	}
	if strings.Contains(SystemPrompt(testInfo, true), marker) {
		t.Fatal("integration prompt must not forbid cd")
	}
}
