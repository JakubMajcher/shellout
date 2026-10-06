// SPDX-License-Identifier: GPL-3.0-or-later

package shellinit

import (
	"strings"
	"testing"
)

// The first-run guide hands over a line the user pastes into a shell, so the
// path must be the file their shell actually reads, and it must stay relative
// to the home directory.
func TestRCPath(t *testing.T) {
	cases := map[string]string{
		"zsh":  "~/.zshrc",
		"bash": "~/.bashrc",
		"fish": "~/.config/fish/config.fish",
	}
	for shell, want := range cases {
		got, err := RCPath(shell)
		if err != nil {
			t.Errorf("%s: %v", shell, err)
			continue
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", shell, got, want)
		}
	}
}

// Printing a home directory into a message someone may paste into a public
// issue is noise, so the tilde form is the only accepted one.
func TestRCPathNeverLeaksHomeDirectory(t *testing.T) {
	for _, shell := range []string{"zsh", "bash", "fish"} {
		got, err := RCPath(shell)
		if err != nil {
			t.Fatal(err)
		}
		if got[0] != '~' {
			t.Errorf("%s: %q does not start with a tilde", shell, got)
		}
		if strings.Contains(got, "/home/") || strings.Contains(got, "/Users/") {
			t.Errorf("%s: %q contains an absolute home directory", shell, got)
		}
	}
}

// /bin/sh is what Collect reports when $SHELL is unset. There is no startup
// file to name for it, and pretending otherwise would produce a wrong path.
func TestRCPathUnknownShell(t *testing.T) {
	for _, shell := range []string{"sh", "tcsh", "", "powershell"} {
		if _, err := RCPath(shell); err == nil {
			t.Errorf("%q: expected an error", shell)
		}
	}
}

// A shell that has an integration must also have a startup file to add it to,
// otherwise the guide would suggest nothing for a shell it supports.
func TestEveryScriptHasAnRCPath(t *testing.T) {
	for _, shell := range []string{"zsh", "bash", "fish"} {
		if _, err := Script(shell); err != nil {
			t.Errorf("Script(%s): %v", shell, err)
		}
		if _, err := RCPath(shell); err != nil {
			t.Errorf("RCPath(%s): %v", shell, err)
		}
	}
}
