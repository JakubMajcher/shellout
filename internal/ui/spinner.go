// SPDX-License-Identifier: GPL-3.0-or-later

package ui

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// StartSpinner draws a spinner on w until stop is called. When enabled is
// false it does nothing. stop clears the spinner line.
func StartSpinner(w io.Writer, enabled bool) (stop func()) {
	if !enabled {
		return func() {}
	}
	done := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		frames := `|/-\`
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for i := 0; ; i++ {
			fmt.Fprintf(w, "\r%c thinking...", frames[i%len(frames)])
			select {
			case <-done:
				fmt.Fprint(w, "\r\x1b[K")
				return
			case <-ticker.C:
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() {
			close(done)
			<-finished
		})
	}
}
