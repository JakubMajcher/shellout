// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/JakubMajcher/shellout/internal/app"
)

func TestParseArgs(t *testing.T) {
	cases := []struct {
		args []string
		want invocation
	}{
		{[]string{"list", "open", "ports"}, invocation{request: "list open ports"}},
		{[]string{"-p", "local", "list", "ports"}, invocation{profile: "local", request: "list ports"}},
		{[]string{"find", "files", "-size", "+1G"}, invocation{request: "find files -size +1G"}},
		{[]string{"init", "zsh"}, invocation{isInit: true, initShell: "zsh"}},
		{[]string{"init", "tcsh"}, invocation{isInit: true, initShell: "tcsh"}},
		{[]string{"init", "a", "git", "repo", "here"}, invocation{request: "init a git repo here"}},
		{[]string{"init"}, invocation{request: "init"}},
		{[]string{"--version"}, invocation{version: true}},
		{[]string{"-h"}, invocation{help: true}},
		{[]string{"--help"}, invocation{help: true}},
		{nil, invocation{}},
	}
	for _, c := range cases {
		got, err := parseArgs(c.args)
		if err != nil {
			t.Errorf("%v: unexpected error %v", c.args, err)
			continue
		}
		if got != c.want {
			t.Errorf("%v: got %+v, want %+v", c.args, got, c.want)
		}
	}
	if _, err := parseArgs([]string{"-x"}); err == nil {
		t.Error("unknown flag must be an error")
	}
	if _, err := parseArgs([]string{"-p"}); err == nil {
		t.Error("-p without a value must be an error")
	}
}

// captureStdout runs fn with os.Stdout replaced by a pipe and returns what was
// written. os.Stdout is swapped because main prints usage and --version to it
// directly. These branches return before any terminal is needed, which is the
// only part of run reachable from a test without a pty.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	return capture(t, &os.Stdout, fn)
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	return capture(t, &os.Stderr, fn)
}

func capture(t *testing.T, target **os.File, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := *target
	*target = w

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()

	*target = old
	w.Close()
	out := <-done
	r.Close()
	return out
}

// TestParseArgsStopsAtFirstNonFlagWord pins Review Focus, case 5: flag parsing
// stops at the first non-flag word, so a request that looks like flags reaches
// the model unchanged instead of being eaten by the flag parser.
func TestParseArgsStopsAtFirstNonFlagWord(t *testing.T) {
	got, err := parseArgs([]string{"list", "files", "-p", "local"})
	if err != nil {
		t.Fatal(err)
	}
	if got.profile != "" {
		t.Errorf("profile = %q, want empty: the -p belongs to the request", got.profile)
	}
	if want := "list files -p local"; got.request != want {
		t.Errorf("request = %q, want %q", got.request, want)
	}
}

// TestParseArgsInitOnlyWithExactlyTwoWords pins Review Focus, case 1: only
// exactly two words, "init <shell>", select init mode. Anything else is a
// request for the model, including a three word phrase starting with init.
func TestParseArgsInitOnlyWithExactlyTwoWords(t *testing.T) {
	cases := []struct {
		args    []string
		isInit  bool
		request string
	}{
		{[]string{"init", "zsh"}, true, ""},
		{[]string{"init", "a", "git", "repo"}, false, "init a git repo"},
		{[]string{"init"}, false, "init"},
		{[]string{"initialise", "zsh"}, false, "initialise zsh"},
	}
	for _, c := range cases {
		got, err := parseArgs(c.args)
		if err != nil {
			t.Errorf("%v: %v", c.args, err)
			continue
		}
		if got.isInit != c.isInit || got.request != c.request {
			t.Errorf("%v: got isInit=%v request=%q, want %v / %q",
				c.args, got.isInit, got.request, c.isInit, c.request)
		}
	}
}

// Section D12 fixes exit codes as part of the contract. These cover the branches
// that return before the terminal check.
func TestRunHelp(t *testing.T) {
	for _, arg := range []string{"-h", "--help"} {
		var code int
		out := captureStdout(t, func() { code = run([]string{arg}) })
		if code != 0 {
			t.Errorf("%s: exit = %d, want 0", arg, code)
		}
		if !strings.Contains(out, "shellout [-p PROFILE] [REQUEST...]") {
			t.Errorf("%s: usage missing from output: %q", arg, out)
		}
		if !strings.Contains(out, "shellout init zsh|bash|fish") {
			t.Errorf("%s: init line missing from usage: %q", arg, out)
		}
	}
}

func TestRunVersion(t *testing.T) {
	var code int
	out := captureStdout(t, func() { code = run([]string{"--version"}) })
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.TrimSpace(out) != app.Version {
		t.Fatalf("printed %q, want %q", strings.TrimSpace(out), app.Version)
	}
}

func TestRunInit(t *testing.T) {
	for _, shell := range []string{"zsh", "bash", "fish"} {
		var code int
		out := captureStdout(t, func() { code = run([]string{"init", shell}) })
		if code != 0 {
			t.Errorf("init %s: exit = %d, want 0", shell, code)
		}
		if !strings.Contains(out, "SHELLOUT_EMIT_FILE=") {
			t.Errorf("init %s: script does not set SHELLOUT_EMIT_FILE: %q", shell, out)
		}
		if !strings.Contains(out, "function shellout") && !strings.Contains(out, "shellout() {") {
			t.Errorf("init %s: script does not define the shellout function", shell)
		}
		if !strings.Contains(out, "sho") {
			t.Errorf("init %s: script does not mention the sho alias", shell)
		}
	}
}

// An unsupported shell is bad usage: exit 2, and both the error and the usage
// go to stderr, because stdout is reserved for --help, --version and init.
func TestRunInitUnsupportedShell(t *testing.T) {
	var code int
	var stdout string
	errOut := captureStderr(t, func() {
		stdout = captureStdout(t, func() { code = run([]string{"init", "tcsh"}) })
	})
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, `unsupported shell "tcsh"`) {
		t.Errorf("stderr = %q", errOut)
	}
	if !strings.HasPrefix(errOut, app.Name+":") {
		t.Errorf("error is not prefixed with the program name: %q", errOut)
	}
	if !strings.Contains(errOut, "Usage:") {
		t.Errorf("usage missing from stderr: %q", errOut)
	}
	if stdout != "" {
		t.Errorf("nothing may go to stdout here, got %q", stdout)
	}
}

// Bad usage exits 2 the same way: message and usage on stderr, stdout untouched.
func TestRunBadUsage(t *testing.T) {
	for _, args := range [][]string{{"-x"}, {"-p"}} {
		var code int
		var stdout string
		errOut := captureStderr(t, func() {
			stdout = captureStdout(t, func() { code = run(args) })
		})
		if code != 2 {
			t.Errorf("%v: exit = %d, want 2", args, code)
		}
		if errOut == "" {
			t.Errorf("%v: nothing on stderr", args)
		}
		if !strings.Contains(errOut, "Usage:") {
			t.Errorf("%v: usage missing from stderr: %q", args, errOut)
		}
		if stdout != "" {
			t.Errorf("%v: stdout must stay empty, got %q", args, stdout)
		}
	}
}

// The usage text must keep describing the commands the README documents.
func TestUsageMentionsEveryCommand(t *testing.T) {
	for _, want := range []string{"-p PROFILE", "REQUEST", "init zsh|bash|fish", "-h", "--version"} {
		if !strings.Contains(usage, want) {
			t.Errorf("usage missing %q", want)
		}
	}
}

// Messages always print the full name, whatever argv[0] was: the binary is also
// installed as the sho symlink (section D2).
func TestFailUsesFullName(t *testing.T) {
	var code int
	out := captureStderr(t, func() { code = fail(1, "boom") })
	if code != 1 {
		t.Fatalf("fail returned %d, want 1", code)
	}
	if out != app.Name+": boom\n" {
		t.Fatalf("fail printed %q, want %q", out, app.Name+": boom\n")
	}
}
