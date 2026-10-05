// SPDX-License-Identifier: GPL-3.0-or-later

package ui

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSpinnerDisabledWritesNothing(t *testing.T) {
	var buf bytes.Buffer
	stop := StartSpinner(&buf, false)
	time.Sleep(150 * time.Millisecond)
	stop()
	if buf.Len() != 0 {
		t.Fatalf("wrote %q", buf.String())
	}
}

func TestSpinnerClearsLineOnStop(t *testing.T) {
	var buf bytes.Buffer
	stop := StartSpinner(&buf, true)
	time.Sleep(150 * time.Millisecond)
	stop()
	stop() // second call must not panic
	if !strings.HasSuffix(buf.String(), "\r\x1b[K") {
		t.Fatalf("line not cleared: %q", buf.String())
	}
}
