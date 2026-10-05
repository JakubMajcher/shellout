// SPDX-License-Identifier: GPL-3.0-or-later

// Package ui draws the suggestion and reads the user's decision on stderr/stdin.
package ui

import "strings"

// Decision is the user's answer to "Run? [y/N/e]".
type Decision int

const (
	Decline Decision = iota
	Approve
	Edit
)

const (
	boldYellow = "\x1b[1;33m"
	bold       = "\x1b[1m"
	reset      = "\x1b[0m"
)

// Decide maps one key press to a decision. Anything but y/Y/e/E declines.
func Decide(key byte) Decision {
	switch key {
	case 'y', 'Y':
		return Approve
	case 'e', 'E':
		return Edit
	default:
		return Decline
	}
}

// ColorEnabled reports whether to use ANSI colors.
func ColorEnabled(isTTY, noColorSet bool) bool {
	return isTTY && !noColorSet
}

// FormatSuggestion renders the description and the indented command.
func FormatSuggestion(desc, cmd string, color bool) string {
	var b strings.Builder
	if color && strings.HasPrefix(desc, "Warning:") {
		b.WriteString(boldYellow + desc + reset)
	} else {
		b.WriteString(desc)
	}
	b.WriteString("\n")
	for _, line := range strings.Split(cmd, "\n") {
		b.WriteString("  ")
		if color {
			b.WriteString(bold + line + reset)
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}
