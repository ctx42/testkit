// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_DefaultImgRmOptions(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// --- When ---
		have := DefaultImgRmOptions()

		// --- Then ---
		assert.NotNil(t, have)
		assert.Duration(t, 250*time.Millisecond, have.sleep)
		assert.Equal(t, 20, have.tries)
		assert.False(t, have.ignoreErrors)
	})

	t.Run("options applied", func(t *testing.T) {
		// --- Given ---
		opt := WithImgRmTries(5)

		// --- When ---
		have := DefaultImgRmOptions(opt)

		// --- Then ---
		assert.NotNil(t, have)
		assert.Duration(t, 250*time.Millisecond, have.sleep)
		assert.Equal(t, 5, have.tries)
		assert.False(t, have.ignoreErrors)
	})
}

func Test_WithImgRmSleep(t *testing.T) {
	// --- Given ---
	opts := &ImgRmOptions{}

	// --- When ---
	WithImgRmSleep(time.Second)(opts)

	// --- Then ---
	assert.Duration(t, time.Second, opts.sleep)
}

func Test_WithImgRmTries(t *testing.T) {
	// --- Given ---
	opts := &ImgRmOptions{}

	// --- When ---
	WithImgRmTries(10)(opts)

	// --- Then ---
	assert.Equal(t, 10, opts.tries)
}

func Test_WithImgRmIgnoreErrors(t *testing.T) {
	// --- Given ---
	opts := &ImgRmOptions{}

	// --- When ---
	WithImgRmIgnoreErrors()(opts)

	// --- Then ---
	assert.True(t, opts.ignoreErrors)
}
