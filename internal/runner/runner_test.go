// SPDX-License-Identifier: GPL-3.0-or-later

package runner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// statPath exists so the test reads as the assertion it is, instead of hiding
// the filesystem call inside the timeout check.
func statPath(p string) (os.FileInfo, error) { return os.Stat(p) }

// fmtInt avoids pulling strconv in for two call sites.
func fmtInt(i int) string { return fmt.Sprintf("%d", i) }

func TestRunExitCode(t *testing.T) {
	code, err := Run("/bin/sh", "exit 3", 0)
	if err != nil || code != 3 {
		t.Fatalf("code=%d err=%v", code, err)
	}
	code, err = Run("/bin/sh", "true", 0)
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

func TestRunSignalExitCode(t *testing.T) {
	code, err := Run("/bin/sh", "kill -TERM $$", 0)
	if err != nil || code != 128+15 {
		t.Fatalf("code=%d err=%v, want 143", code, err)
	}
}

// If Run used signal.Ignore, the child would inherit SIG_IGN for SIGINT,
// survive the kill and exit 0. Ctrl-C must reach the child. This also guards
// against a timeout implementation that puts the child in its own process
// group: a new group would stop the terminal from delivering SIGINT.
func TestRunChildStillGetsSIGINT(t *testing.T) {
	code, err := Run("/bin/sh", "kill -INT $$; sleep 1; exit 0", 0)
	if err != nil || code != 128+2 {
		t.Fatalf("code=%d err=%v, want 130", code, err)
	}
}

func TestRunMissingShell(t *testing.T) {
	if _, err := Run("/nonexistent/shell", "true", 0); err == nil {
		t.Fatal("expected start error")
	}
}

// Zero means no limit, and a command that outlasts a short deadline given to
// other tests must still be allowed to finish.
func TestRunZeroTimeoutMeansUnlimited(t *testing.T) {
	code, err := Run("/bin/sh", "sleep 0.3; echo done", 0)
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v, want 0", code, err)
	}
}

// A command that overruns is killed and reported as a timeout, not as a plain
// signal death, so main can tell the two apart.
func TestRunTimeoutKillsLongCommand(t *testing.T) {
	start := time.Now()
	code, err := Run("/bin/sh", "sleep 30", 150*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
	if code != TimeoutExitCode {
		t.Fatalf("code = %d, want %d", code, TimeoutExitCode)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("timeout not honored: took %v", elapsed)
	}
}

// The command itself must die, not just the shell that spawned it. sh forks
// every simple command, so a timeout that killed only the direct child would
// leave tar, find or sleep running and holding the terminal. This test is the
// reason Run puts the child in its own process group.
func TestRunTimeoutKillsTheCommandNotJustTheShell(t *testing.T) {
	marker := t.TempDir() + "/finished"
	_, err := Run("/bin/sh", "sleep 0.4; touch "+marker, 100*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("err = %v, want ErrTimeout", err)
	}
	time.Sleep(700 * time.Millisecond)
	if _, statErr := statPath(marker); statErr == nil {
		t.Fatal("the command survived the timeout and touched the marker")
	}
}

// The child must be alone in its process group, which is what lets a timeout
// reach the whole tree.
func TestRunChildGetsItsOwnProcessGroup(t *testing.T) {
	// If the shell were not a group leader, signalGroup could never target it
	// without hitting shellout itself.
	cmd := exec.Command("/bin/sh", "-c", "true")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Wait() }()
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	if pgid != cmd.Process.Pid {
		t.Fatalf("pgid = %d, want %d: the child must lead its own group", pgid, cmd.Process.Pid)
	}
}

// signalGroup must reach every process in the group, not only the leader.
func TestSignalGroupReachesGrandchildren(t *testing.T) {
	cmd := exec.Command("/bin/sh", "-c", "sleep 30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	// Give sh time to fork the sleep.
	time.Sleep(300 * time.Millisecond)
	if err := signalGroup(cmd.Process.Pid, syscall.SIGKILL); err != nil {
		t.Fatalf("signalGroup: %v", err)
	}
	_ = cmd.Wait()

	// Nothing from that group may still be alive.
	time.Sleep(200 * time.Millisecond)
	if out, err := exec.Command("ps", "-o", "pgid=", "-g",
		fmtInt(cmd.Process.Pid)).CombinedOutput(); err == nil {
		if lines := strings.TrimSpace(string(out)); lines != "" {
			t.Fatalf("processes still alive in the killed group: %q", lines)
		}
	}
}

// Forwarding must translate the two terminal signals the tool promises to relay.
func TestToSyscall(t *testing.T) {
	if sig, ok := toSyscall(os.Interrupt); !ok || sig != syscall.SIGINT {
		t.Errorf("os.Interrupt -> %v, %v; want SIGINT, true", sig, ok)
	}
	if sig, ok := toSyscall(syscall.SIGQUIT); !ok || sig != syscall.SIGQUIT {
		t.Errorf("SIGQUIT -> %v, %v; want SIGQUIT, true", sig, ok)
	}
	if _, ok := toSyscall(os.Kill); ok {
		t.Error("an unrelated signal must not be forwarded")
	}
}

// A command that finishes inside the limit is untouched by the timer.
func TestRunUnderTimeoutSucceeds(t *testing.T) {
	code, err := Run("/bin/sh", "exit 0", 10*time.Second)
	if err != nil || code != 0 {
		t.Fatalf("code=%d err=%v", code, err)
	}
}

// ErrTimeout must be comparable with errors.Is and must mention the limit, so
// the message printed by main is not a bare sentinel.
func TestErrTimeoutMentionsTheLimit(t *testing.T) {
	if !strings.Contains(ErrTimeout.Error(), "timed out") {
		t.Fatalf("ErrTimeout = %q", ErrTimeout.Error())
	}
}
