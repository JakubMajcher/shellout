// SPDX-License-Identifier: GPL-3.0-or-later

package ui

import (
	"strings"
	"testing"
)

// Pressing e opens a line that is already filled in. If that line wore the same
// prompt as the request line, the two are indistinguishable and the pre-filled
// command looks like the tool is waiting for a request instead of an edit.
func TestEditPromptDiffersFromRequestPrompt(t *testing.T) {
	if editPrompt == requestPrompt {
		t.Fatalf("edit prompt %q equals request prompt %q", editPrompt, requestPrompt)
	}
	if !strings.Contains(editPrompt, "edit") {
		t.Errorf("edit prompt %q does not say it is an edit", editPrompt)
	}
	if !strings.HasPrefix(requestPrompt, ">") {
		t.Errorf("request prompt = %q, want the plain > prompt", requestPrompt)
	}
}

// Both prompts must end with a space, or the cursor sits on top of the
// pre-filled text.
func TestPromptsEndWithSpace(t *testing.T) {
	for name, p := range map[string]string{"request": requestPrompt, "edit": editPrompt} {
		if !strings.HasSuffix(p, " ") {
			t.Errorf("%s prompt %q does not end with a space", name, p)
		}
	}
}
