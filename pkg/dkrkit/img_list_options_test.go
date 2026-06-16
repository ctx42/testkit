// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_WithImgLsFilter(t *testing.T) {
	t.Run("single filter", func(t *testing.T) {
		// --- Given ---
		opts := &ImgListOptions{}

		// --- When ---
		WithImgLsFilter("reference=my-img")(opts)

		// --- Then ---
		assert.Equal(t, []string{"reference=my-img"}, opts.filters)
	})

	t.Run("multiple filters in one call", func(t *testing.T) {
		// --- Given ---
		opts := &ImgListOptions{}

		// --- When ---
		WithImgLsFilter("reference=my-img", "dangling=false")(opts)

		// --- Then ---
		want := []string{"reference=my-img", "dangling=false"}
		assert.Equal(t, want, opts.filters)
	})

	t.Run("appends on repeated calls", func(t *testing.T) {
		// --- Given ---
		opts := &ImgListOptions{}

		// --- When ---
		WithImgLsFilter("reference=my-img")(opts)
		WithImgLsFilter("dangling=false")(opts)

		// --- Then ---
		want := []string{"reference=my-img", "dangling=false"}
		assert.Equal(t, want, opts.filters)
	})
}
