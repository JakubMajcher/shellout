// SPDX-License-Identifier: GPL-3.0-or-later

// Package runner runs the approved command through the user's shell.
package runner

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// Run executes `shellPath -c command` with the terminal attached.
// While the child runs, SIGINT and SIGQUIT are caught and dropped here, so
// they only affect the child. Catching (signal.Notify) instead of ignoring
// matters: ignored signals are inherited across exec, caught ones are reset.
func Run(shellPath, command string) (int, error) {
	cmd := exec.Command(shellPath, "-c", command)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGQUIT)
	defer signal.Stop(sigs)

	err := cmd.Run()
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
