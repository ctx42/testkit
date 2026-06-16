// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package testkit

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Wait4File(t *testing.T) {
	t.Run("file created within timeout", func(t *testing.T) {
		// --- Given ---
		var err error
		var data string
		started := make(chan struct{})
		done := make(chan struct{})
		pth := filepath.Join(t.TempDir(), "file.txt")

		// --- When ---
		go func() {
			close(started)
			data, err = Wait4File("1s", pth)
			close(done)
		}()

		// --- Then ---
		<-started
		time.Sleep(100 * time.Millisecond)
		assert.NoError(t, os.WriteFile(pth, []byte("content"), 0600))
		<-done
		assert.Equal(t, "content", data)
		assert.NoError(t, err)
	})

	t.Run("the file isn't created within timeout", func(t *testing.T) {
		// --- Given ---
		var err error
		var data string
		started := make(chan struct{})
		done := make(chan struct{})
		pth := filepath.Join(t.TempDir(), "file.txt")

		// --- When ---
		go func() {
			close(started)
			data, err = Wait4File("1ms", pth)
			close(done)
		}()

		// --- Then ---
		<-started
		<-done
		assert.Empty(t, data)
		wMsg := "" +
			"timeout waiting for file read:\n" +
			"    within: 1ms\n" +
			"  throttle: 50ms\n" +
			"      file: %s"
		assert.ErrorEqual(t, fmt.Sprintf(wMsg, pth), err)
	})

	t.Run("file created within timeout empty first", func(t *testing.T) {
		// --- Given ---
		var err error
		var data string
		started := make(chan struct{})
		done := make(chan struct{})
		pth := filepath.Join(t.TempDir(), "file.txt")

		// --- When ---
		go func() {
			close(started)
			data, err = Wait4File("1s", pth)
			close(done)
		}()

		// --- Then ---
		<-started
		time.Sleep(100 * time.Millisecond)
		assert.NoError(t, os.WriteFile(pth, nil, 0600))
		time.Sleep(10 * time.Millisecond)
		assert.NoError(t, os.WriteFile(pth, []byte("content"), 0600))
		<-done
		assert.Equal(t, "content", data)
		assert.NoError(t, err)
	})
}
