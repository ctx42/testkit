// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_DefaultCtrRunOptions(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// --- When ---
		have := DefaultCtrRunOptions()

		// --- Then ---
		assert.NotNil(t, have)
		assert.Nil(t, have.args)
		assert.Nil(t, have.labels)
		assert.Equal(t, "", have.cidPth)
		assert.Nil(t, have.cidCh)
		assert.Equal(t, 50*time.Millisecond, have.cidPoll)
		assert.False(t, have.detach)
		assert.True(t, have.remove)
	})

	t.Run("options applied", func(t *testing.T) {
		// --- Given ---
		opt := WithCtrRunArgs("a", "b")

		// --- When ---
		have := DefaultCtrRunOptions(opt)

		// --- Then ---
		assert.NotNil(t, have)
		assert.Equal(t, []string{"a", "b"}, have.args)
		assert.Nil(t, have.labels)
		assert.Equal(t, "", have.cidPth)
		assert.Nil(t, have.cidCh)
		assert.Equal(t, 50*time.Millisecond, have.cidPoll)
		assert.False(t, have.detach)
		assert.True(t, have.remove)
	})
}

func Test_WithCtrRunArgs(t *testing.T) {
	t.Run("appends to nil", func(t *testing.T) {
		// --- Given ---
		opts := &CtrRunOptions{}

		// --- When ---
		WithCtrRunArgs("a", "b", "c")(opts)

		// --- Then ---
		assert.Equal(t, []string{"a", "b", "c"}, opts.args)
	})

	t.Run("accumulates across calls", func(t *testing.T) {
		// --- Given ---
		opts := &CtrRunOptions{}

		// --- When ---
		WithCtrRunArgs("a", "b")(opts)
		WithCtrRunArgs("c")(opts)

		// --- Then ---
		assert.Equal(t, []string{"a", "b", "c"}, opts.args)
	})
}

func Test_WithCtrRunLabel(t *testing.T) {
	// --- Given ---
	opts := &CtrRunOptions{}

	// --- When ---
	WithCtrRunLabel("key", "val")(opts)

	// --- Then ---
	assert.HasKeyValue(t, "key", "val", opts.labels)
}

func Test_WithCtrRunLabels(t *testing.T) {
	// --- Given ---
	opts := &CtrRunOptions{}
	labels := map[string]string{"LAB0": "VAL0", "LAB1": "VAL1"}

	// --- When ---
	WithCtrRunLabels(labels)(opts)

	// --- Then ---
	want := map[string]string{"LAB0": "VAL0", "LAB1": "VAL1"}
	assert.Equal(t, want, opts.labels)
	want["LAB2"] = "VAL3"
	assert.NotEqual(t, want, opts.labels)
}

func Test_WithCtrRunCIDPth(t *testing.T) {
	// --- Given ---
	opts := &CtrRunOptions{}

	// --- When ---
	WithCtrRunCIDPth("/dir/file.log")(opts)

	// --- Then ---
	assert.Equal(t, "/dir/file.log", opts.cidPth)
}

func Test_WithCtrRunCID(t *testing.T) {
	// --- Given ---
	opts := &CtrRunOptions{}

	// --- When ---
	ch, opt := WithCtrRunCID()

	// --- Then ---
	assert.NotNil(t, ch)
	opt(opts)
	opts.cidCh <- "test"
	close(opts.cidCh)
	assert.Equal(t, "test", <-ch)
	assert.Equal(t, "", opts.cidPth)
}

func Test_WithCtrRunDetach(t *testing.T) {
	// --- Given ---
	opts := &CtrRunOptions{}

	// --- When ---
	WithCtrRunDetach()(opts)

	// --- Then ---
	assert.True(t, opts.detach)
}

func Test_WithCtrRunNoRemove(t *testing.T) {
	// --- Given ---
	opts := &CtrRunOptions{remove: true}

	// --- When ---
	WithCtrRunNoRemove()(opts)

	// --- Then ---
	assert.False(t, opts.remove)
}
