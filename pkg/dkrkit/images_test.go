// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Images_FindByRef(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		// --- Given ---
		img0 := &Image{ID: "id0", Repository: "rep0", Tag: "tag0"}
		img1 := &Image{ID: "id1", Repository: "rep1", Tag: "tag1"}
		img2 := &Image{ID: "id2", Repository: "rep2", Tag: "tag2"}

		ims := Images{img0, img1, img2}

		// --- When ---
		have := ims.FindByRef("rep1:tag1")

		// --- Then ---
		assert.Same(t, img1, have)
	})

	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		img0 := &Image{ID: "id0", Repository: "rep0", Tag: "tag0"}
		img1 := &Image{ID: "id1", Repository: "rep1", Tag: "tag1"}
		img2 := &Image{ID: "id2", Repository: "rep2", Tag: "tag2"}

		ims := Images{img0, img1, img2}

		// --- When ---
		have := ims.FindByRef("rep3:tag")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		ims := Images{}

		// --- When ---
		have := ims.FindByRef("rep0:tag0")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_Images_FindByID(t *testing.T) {
	t.Run("found with hash name", func(t *testing.T) {
		// --- Given ---
		img0 := &Image{ID: "sha256:id0", Repository: "rep0", Tag: "tag0"}
		img1 := &Image{ID: "sha256:id1", Repository: "rep1", Tag: "tag1"}
		img2 := &Image{ID: "sha256:id2", Repository: "rep2", Tag: "tag2"}

		ims := Images{img0, img1, img2}

		// --- When ---
		have := ims.FindByID("sha256:id1")

		// --- Then ---
		assert.Same(t, img1, have)
	})

	t.Run("found without a hash name", func(t *testing.T) {
		// --- Given ---
		img0 := &Image{ID: "sha256:id0", Repository: "rep0", Tag: "tag0"}
		img1 := &Image{ID: "sha256:id1", Repository: "rep1", Tag: "tag1"}
		img2 := &Image{ID: "sha256:id2", Repository: "rep2", Tag: "tag2"}

		ims := Images{img0, img1, img2}

		// --- When ---
		have := ims.FindByID("id1")

		// --- Then ---
		assert.Same(t, img1, have)
	})

	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		img0 := &Image{ID: "id0", Repository: "rep0", Tag: "tag0"}
		img1 := &Image{ID: "id1", Repository: "rep1", Tag: "tag1"}
		img2 := &Image{ID: "id2", Repository: "rep2", Tag: "tag2"}

		ims := Images{img0, img1, img2}

		// --- When ---
		have := ims.FindByID("id3")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		ims := Images{}

		// --- When ---
		have := ims.FindByID("id3")

		// --- Then ---
		assert.Nil(t, have)
	})
}
