// SPDX-License-Identifier: GPL-3.0-or-later

package runner

import "testing"

func TestRunExitCode(t *testing.T) {
	code, err := Run("/bin/sh", "exit 3")
	if err != nil || code != 3 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	code, err = Run("/bin/sh", "true")
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestRunSignalExitCode(t *testing.T) {
	code, err := Run("/bin/sh", "kill -TERM $$")
	if err != nil || code != 128+15 {
		t.Fatalf("code=%d err=%v, want 143", code, err)
	}
}

// If Run used signal.Ignore, the child would inherit SIG_IGN for SIGINT,
// survive the kill and exit 0. Ctrl-C must reach the child.
func TestRunChildStillGetsSIGINT(t *testing.T) {
	code, err := Run("/bin/sh", "kill -INT $$; sleep 1; exit 0")
	if err != nil || code != 128+2 {
		t.Fatalf("code=%d err=%v, want 130", code, err)
	}
}

func TestRunMissingShell(t *testing.T) {
	if _, err := Run("/nonexistent/shell", "true"); err == nil {
		t.Fatal("expected start error")
	}
}
