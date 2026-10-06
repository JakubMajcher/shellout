// SPDX-License-Identifier: GPL-3.0-or-later

// Package runner runs the approved command through the user's shell.
package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// TimeoutExitCode is what Run returns when the command outran its limit. It is
// the same code GNU timeout(1) uses, so scripts already know it.
const TimeoutExitCode = 124

// ErrTimeout means the command was killed for running too long. Compare with
// errors.Is; the message carries the limit so main can print it as is.
var ErrTimeout = errors.New("command timed out")

// Run executes `shellPath -c command` with the terminal attached.
//
// While the child runs, SIGINT and SIGQUIT are caught here and forwarded to the
// child instead of being left to the default disposition. Catching
// (signal.Notify) rather than ignoring matters twice over: ignored signals are
// inherited across exec, and caught ones can be forwarded.
//
// A non-zero limit kills the whole command when it overruns and returns
// TimeoutExitCode with ErrTimeout. Zero means no limit, which is the default on
// purpose: a limit that fires mid-command can leave a half-written archive or a
// partial transfer behind, and the user already approved this command and can
// see that it is long.
//
// The child gets its own process group. Without one a timeout could only kill
// the shell, because sh forks every simple command instead of execing it, so
// the real command would survive as an orphan and keep holding the terminal.
// The cost is that the terminal no longer delivers signals to the child
// directly, which is why SIGINT and SIGQUIT are forwarded by hand. SIGTSTP is
// not forwarded, so Ctrl-Z suspends shellout rather than the command.
func Run(shellPath, command string, limit time.Duration) (int, error) {
	ctx := context.Background()
	if limit > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, limit)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, shellPath, "-c", command)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// Kill the group, not just the shell, so a timed out pipeline leaves
	// nothing behind. The group id equals the shell's pid because of Setpgid.
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return signalGroup(cmd.Process.Pid, syscall.SIGKILL)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGQUIT)
	defer signal.Stop(sigs)
	done := make(chan struct{})
	defer close(done)
	go forwardSignals(sigs, cmd, done)

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return TimeoutExitCode, fmt.Errorf("%w after %s", ErrTimeout, limit)
	}
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			return 128 + int(ws.Signal()), nil
		}
		return exitErr.ExitCode(), nil
	}
	return 0, err
}

// forwardSignals relays terminal signals to the child's process group. It stops
// when done is closed, so the goroutine cannot outlive the command.
func forwardSignals(sigs <-chan os.Signal, cmd *exec.Cmd, done <-chan struct{}) {
	for {
		select {
		case s := <-sigs:
			if cmd.Process == nil {
				continue
			}
			sig, ok := toSyscall(s)
			if !ok {
				continue
			}
			_ = signalGroup(cmd.Process.Pid, sig)
		case <-done:
			return
		}
	}
}

// signalGroup sends sig to the whole process group led by pid. The negative pid
// is what makes this reach the grandchildren too.
func signalGroup(pid int, sig syscall.Signal) error {
	return syscall.Kill(-pid, sig)
}

func toSyscall(s os.Signal) (syscall.Signal, bool) {
	switch s {
	case os.Interrupt:
		return syscall.SIGINT, true
	case syscall.SIGQUIT:
		return syscall.SIGQUIT, true
	}
	return 0, false
}
