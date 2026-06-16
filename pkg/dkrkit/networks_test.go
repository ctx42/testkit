// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Networks_FindByName(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		// --- Given ---
		net0 := &Network{ID: "id0", Name: "name0"}
		net1 := &Network{ID: "id1", Name: "name1"}
		net2 := &Network{ID: "id2", Name: "name2"}

		nts := Networks{net0, net1, net2}

		// --- When ---
		have := nts.FindByName("name1")

		// --- Then ---
		assert.Same(t, net1, have)
	})

	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		net0 := &Network{ID: "id0", Name: "name0"}
		net1 := &Network{ID: "id1", Name: "name1"}
		net2 := &Network{ID: "id2", Name: "name2"}

		nts := Networks{net0, net1, net2}

		// --- When ---
		have := nts.FindByName("name3")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		nts := Networks{}

		// --- When ---
		have := nts.FindByName("name1")

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_Networks_FindByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		// --- Given ---
		net0 := &Network{ID: "id0", Name: "name0"}
		net1 := &Network{ID: "id1", Name: "name1"}
		net2 := &Network{ID: "id2", Name: "name2"}

		nts := Networks{net0, net1, net2}

		// --- When ---
		have := nts.FindByID("id1")

		// --- Then ---
		assert.Same(t, net1, have)
	})

	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		net0 := &Network{ID: "id0", Name: "name0"}
		net1 := &Network{ID: "id1", Name: "name1"}
		net2 := &Network{ID: "id2", Name: "name2"}

		nts := Networks{net0, net1, net2}

		// --- When ---
		have := nts.FindByID("id3")

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		net0 := &Network{ID: "id0", Name: "name0"}
		net1 := &Network{ID: "id1", Name: "name1"}
		net2 := &Network{ID: "id2", Name: "name2"}

		nts := Networks{net0, net1, net2}

		// --- When ---
		have := nts.FindByID("id3")

		// --- Then ---
		assert.Nil(t, have)
	})
}
