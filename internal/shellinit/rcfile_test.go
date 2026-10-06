package shellinit

import (
	"testing"
)

// The first-run guide hands over a ready to paste line, so the path must be the
// file the user's shell actually reads.
func TestRCFile(t *testing.T) {
	cases := map[string]string{
		"zsh":  "/home/u/.zshrc",
		"bash": "/home/u/.bashrc",
		"fish": "/home/u/.config/fish/config.fish",
	}
	for shell, want := range cases {
		got, err := RCFile(shell, "/home/u")
		if err != nil {
			t.Errorf("%s: %v", shell, err)
			continue
		}
		if got != want {
			t.Errorf("%s: got %q, want %q", shell, got, want)
		}
	}
}

// /bin/sh is what Collect reports when $SHELL is unset. There is no startup
// file to name for it, and pretending otherwise would produce a wrong path.
func TestRCFileUnknownShell(t *testing.T) {
	if _, err := RCFile("sh", "/home/u"); err == nil {
		t.Fatal("expected an error for sh")
	}
	if _, err := RCFile("tcsh", "/home/u"); err == nil {
		t.Fatal("expected an error for tcsh")
	}
}

// A profile that has no startup file must still get a usable command.
func TestRCFileAndScriptAgreeOnSupportedShells(t *testing.T) {
	for _, shell := range []string{"zsh", "bash", "fish"} {
		if _, err := Script(shell); err != nil {
			t.Errorf("Script(%s): %v", shell, err)
		}
		if _, err := RCFile(shell, "/home/u"); err != nil {
			t.Errorf("RCFile(%s): %v", shell, err)
		}
	}

}
