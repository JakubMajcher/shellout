// SPDX-License-Identifier: GPL-3.0-or-later

package ui

import (
	"strings"
	"testing"
)

func TestDecide(t *testing.T) {
	cases := map[byte]Decision{
		'y': Approve, 'Y': Approve,
		'e': Edit, 'E': Edit,
		'n': Decline, '\r': Decline, '\n': Decline, 27: Decline, 3: Decline, 'x': Decline,
	}
	for key, want := range cases {
		if got := Decide(key); got != want {
			t.Errorf("Decide(%q) = %v, want %v", key, got, want)
		}
	}
}

func TestColorEnabled(t *testing.T) {
	if !ColorEnabled(true, false) || ColorEnabled(true, true) || ColorEnabled(false, false) {
		t.Fatal("ColorEnabled truth table broken")
	}
}

func TestFormatSuggestionPlain(t *testing.T) {
	got := FormatSuggestion("Lists files.", "ls -la", false)
	if got != "Lists files.\n  ls -la\n" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatSuggestionWarningColor(t *testing.T) {
	got := FormatSuggestion("Warning: deletes logs.", "rm -rf logs", true)
	if !strings.HasPrefix(got, "\x1b[1;33mWarning: deletes logs.\x1b[0m\n") {
		t.Fatalf("warning not bold yellow: %q", got)
	}
	if !strings.Contains(got, "  \x1b[1mrm -rf logs\x1b[0m\n") {
		t.Fatalf("command not bold: %q", got)
	}
	plain := FormatSuggestion("Warning: deletes logs.", "rm -rf logs", false)
	if strings.Contains(plain, "\x1b[") {
		t.Fatalf("escape codes without color: %q", plain)
	}
}

func TestFormatSuggestionMultiLine(t *testing.T) {
	got := FormatSuggestion("Two steps.", "cd /tmp &&\n  ls", false)
	if got != "Two steps.\n  cd /tmp &&\n    ls\n" {
		t.Fatalf("got %q", got)
	}
}
