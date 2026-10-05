// SPDX-License-Identifier: GPL-3.0-or-later

package ui

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ergochat/readline"
	"golang.org/x/term"
)

const (
	// requestPrompt asks for the natural-language request.
	requestPrompt = "> "
	// editPrompt is distinct from requestPrompt on purpose. With one prompt for
	// both, the line that opens after pressing e looks exactly like a fresh
	// request, which reads as "shellout is waiting for something" instead of
	// "edit the command that is already here".
	editPrompt = "edit> "
)

// Terminal reads from In (the TTY) and draws on Out (stderr).
type Terminal struct {
	In  *os.File
	Out *os.File
}

// ReadRequest shows a "> " prompt and returns the trimmed line.
// Ctrl-C or Ctrl-D returns "" and a nil error.
func (t Terminal) ReadRequest() (string, error) {
	return t.readLine(requestPrompt, "")
}

// Confirm shows the suggestion and asks "Run? [y/N/e]". It returns the final
// command (edited, if the user chose e) and whether to run it.
func (t Terminal) Confirm(desc, cmd string) (string, bool, error) {
	_, noColor := os.LookupEnv("NO_COLOR")
	color := ColorEnabled(term.IsTerminal(int(t.Out.Fd())), noColor)
	fmt.Fprint(t.Out, FormatSuggestion(desc, cmd, color))
	fmt.Fprint(t.Out, "Run? [y/N/e] ")

	key, err := t.readKey()
	fmt.Fprintln(t.Out)
	if err != nil {
		return "", false, err
	}
	switch Decide(key) {
	case Approve:
		return cmd, true, nil
	case Edit:
		edited, err := t.readLine(editPrompt, cmd)
		if err != nil || edited == "" {
			return "", false, err
		}
		return edited, true, nil
	default:
		return "", false, nil
	}
}

// readKey reads one byte in raw mode and always restores the terminal.
func (t Terminal) readKey() (byte, error) {
	fd := int(t.In.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}
	defer term.Restore(fd, old)
	var buf [1]byte
	if _, err := t.In.Read(buf[:]); err != nil {
		return 0, err
	}
	return buf[0], nil
}

// readLine shows prompt pre-filled with def. Ctrl-C or Ctrl-D returns "".
func (t Terminal) readLine(prompt, def string) (string, error) {
	rl, err := readline.NewFromConfig(&readline.Config{
		Prompt: prompt,
		Stdin:  t.In,
		Stdout: t.Out,
		Stderr: t.Out,
	})
	if err != nil {
		return "", err
	}
	defer rl.Close()
	line, err := rl.ReadLineWithDefault(def)
	if errors.Is(err, readline.ErrInterrupt) || errors.Is(err, io.EOF) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
