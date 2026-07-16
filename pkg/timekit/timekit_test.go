// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package timekit

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

var (
	past   = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	future = time.Now().Add(time.Hour).Truncate(time.Second)
)

func Test_ClockStartingAt(t *testing.T) {
	t.Run("past", func(t *testing.T) {
		// --- Given ---
		clk := ClockStartingAt(past)

		// --- When ---
		have1 := clk()
		have2 := clk()

		// --- Then ---
		assert.True(t, have1.Before(have2))
		assert.True(t, have2.Before(time.Now()))
	})

	t.Run("future", func(t *testing.T) {
		// --- Given ---
		clk := ClockStartingAt(future)

		// --- When ---
		have1 := clk()
		have2 := clk()

		// --- Then ---
		assert.True(t, have1.Before(have2))
		assert.True(t, have2.After(time.Now()))
	})
}

func Test_ClockFixed(t *testing.T) {
	// --- Given ---
	want := time.Now()
	clk := ClockFixed(want)

	// --- When ---
	have0 := clk()
	have1 := clk()
	have2 := clk()

	// --- Then ---
	assert.Equal(t, want, have0)
	assert.Equal(t, want, have1)
	assert.Equal(t, want, have2)
}

func Test_ClockDeterministic(t *testing.T) {
	// --- Given ---
	clk := ClockDeterministic(past, time.Second)

	// --- When ---
	have0 := clk()
	have1 := clk()
	have2 := clk()

	// --- Then ---
	assert.Equal(t, past.Add(0*time.Second), have0)
	assert.Equal(t, past.Add(1*time.Second), have1)
	assert.Equal(t, past.Add(2*time.Second), have2)
}

func Test_TikTak(t *testing.T) {
	// --- Given ---
	clk := TikTak(past)

	// --- When ---
	have0 := clk()
	have1 := clk()
	have2 := clk()

	// --- Then ---
	assert.Equal(t, past.Add(0*time.Second), have0)
	assert.Equal(t, past.Add(1*time.Second), have1)
	assert.Equal(t, past.Add(2*time.Second), have2)
}
