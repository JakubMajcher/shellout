// SPDX-License-Identifier: GPL-3.0-or-later

package shellinit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func needShell(t *testing.T, name string) string {
	t.Helper()
	p, err := exec.LookPath(name)
	if err != nil {
		if os.Getenv("SHELLOUT_REQUIRE_SHELLS") == "1" {
			t.Fatalf("%s is required but not installed", name)
		}
		t.Skipf("%s not installed", name)
	}
	return p
}

func TestScriptUnknownShell(t *testing.T) {
	if _, err := Script("tcsh"); err == nil {
		t.Fatal("expected error for tcsh")
	}
}

func TestScriptsSyntax(t *testing.T) {
	checks := map[string][]string{
		"zsh":  {"-n"},
		"bash": {"-n"},
		"fish": {"--no-execute"},
	}
	for shell, flags := range checks {
		t.Run(shell, func(t *testing.T) {
			bin := needShell(t, shell)
			s, err := Script(shell)
			if err != nil {
				t.Fatal(err)
			}
			f := filepath.Join(t.TempDir(), "init")
			os.WriteFile(f, []byte(s), 0o644)
			out, err := exec.Command(bin, append(flags, f)...).CombinedOutput()
			if err != nil {
				t.Fatalf("syntax error: %v\n%s", err, out)
			}
		})
	}
}

// fakeBinary puts a fake `shellout` first in PATH. It writes `cd '<dir>'` to
// $SHELLOUT_EMIT_FILE and exits 0, like the real binary after approval.
func fakeBinary(t *testing.T) (pathEnv, target string) {
	t.Helper()
	target, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s' \"cd '%s'\" > \"$SHELLOUT_EMIT_FILE\"\nexit 0\n", target)
	if err := os.WriteFile(filepath.Join(binDir, "shellout"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return "PATH=" + binDir + ":" + os.Getenv("PATH"), target
}

func runShell(t *testing.T, bin string, args []string, pathEnv string) string {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), pathEnv, "HOME="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	return string(out)
}

func writeInit(t *testing.T, shell string) string {
	t.Helper()
	s, err := Script(shell)
	if err != nil {
		t.Fatal(err)
	}
	f := filepath.Join(t.TempDir(), "init."+shell)
	os.WriteFile(f, []byte(s), 0o644)
	return f
}

func TestZshIntegration(t *testing.T) {
	bin := needShell(t, "zsh")
	pathEnv, target := fakeBinary(t)
	init := writeInit(t, "zsh")
	// ${history[1]} is zsh's newest history entry, newest first. It is used
	// instead of `fc -ln -1`, which in a non-interactive zsh prints the whole
	// history rather than the last event, so it cannot tell the two apart.
	script := fmt.Sprintf("source %s; cd /; shellout x; pwd; print -r -- ${history[1]}; cd /; sho y; pwd; print -r -- ${history[1]}", init)
	out := runShell(t, bin, []string{"-f", "-c", script}, pathEnv)
	want := fmt.Sprintf("%[1]s\ncd '%[1]s'\n%[1]s\ncd '%[1]s'\n", target)
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
}

// The approved command must be the newest history entry and must not be
// crowded out by the wrapper call itself.
func TestZshHistoryHoldsOnlyRealCommands(t *testing.T) {
	bin := needShell(t, "zsh")
	pathEnv, target := fakeBinary(t)
	init := writeInit(t, "zsh")
	script := fmt.Sprintf("source %s; cd /; shellout x; cd /; sho y; fc -ln", init)
	out := runShell(t, bin, []string{"-f", "-c", script}, pathEnv)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	want := fmt.Sprintf("cd '%s'\ncd '%s'\n", target, target)
	if out != want {
		t.Fatalf("history = %q, want %q", out, want)
	}
	if len(lines) != 2 {
		t.Fatalf("history entries = %d, want 2", len(lines))
	}
}

func TestBashIntegration(t *testing.T) {
	bin := needShell(t, "bash")
	pathEnv, target := fakeBinary(t)
	init := writeInit(t, "bash")
	script := fmt.Sprintf("set -o history; source %s; cd /; shellout x; pwd; history 1; cd /; sho y; pwd; history 1", init)
	out := runShell(t, bin, []string{"--norc", "--noprofile", "-c", script}, pathEnv)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 4 || lines[0] != target || lines[2] != target ||
		!strings.HasSuffix(lines[1], "cd '"+target+"'") || !strings.HasSuffix(lines[3], "cd '"+target+"'") {
		t.Fatalf("unexpected output:\n%s", out)
	}
}

func TestFishIntegration(t *testing.T) {
	bin := needShell(t, "fish")
	pathEnv, target := fakeBinary(t)
	init := writeInit(t, "fish")
	script := fmt.Sprintf(`source %s; cd /; shellout x; pwd; cd /; sho y; pwd
if test (string split . -- $version)[1] -ge 4
    history search --max 1 --prefix "cd '"
end`, init)
	out := runShell(t, bin, []string{"--no-config", "-c", script}, pathEnv)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 || lines[0] != target || lines[1] != target {
		t.Fatalf("unexpected output:\n%s", out)
	}
	if len(lines) == 3 && lines[2] != "cd '"+target+"'" {
		t.Fatalf("history entry = %q", lines[2])
	}
}
