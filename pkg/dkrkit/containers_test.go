// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Container_UnmarshalJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		data := []byte(`{
            "ID": "ID",
            "Created": "2000-01-02T3:04:05Z",
            "Config": {
                "Image": "Image",
                "Env": [
                    "KEY0=VAL0",
                    "KEY1=VAL1"
                ],
                "Labels": {
                    "LAB0": "LV0",
                    "LAB1": "LV1"
                }
            },
            "NetworkSettings": {
                "Networks": {
                    "11-22": { "NetworkID": "AA-BB" },
                    "33-44": { "NetworkID": "CC-DD" }
                }
            }
        }`)

		ctr := &Container{}

		// --- When ---
		err := ctr.UnmarshalJSON(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "ID", ctr.ID)
		assert.Equal(t, "Image", ctr.Image)
		assert.Time(t, "2000-01-02T3:04:05Z", ctr.Created)
		want := map[string]string{
			"LAB0": "LV0",
			"LAB1": "LV1",
		}
		assert.Equal(t, want, ctr.Labels)
		want = map[string]string{
			"KEY0": "VAL0",
			"KEY1": "VAL1",
		}
		assert.Equal(t, want, ctr.Env)
		want = map[string]string{
			"11-22": "AA-BB",
			"33-44": "CC-DD",
		}
		assert.Equal(t, want, ctr.Networks)
	})

	t.Run("nil Container", func(t *testing.T) {
		// --- Given ---
		data := []byte(`{
            "ID": "ID",
            "Created": "2000-01-02T3:04:05Z",
            "Config": { "Image": "Image" }
        }`)

		var ctr Container

		// --- When ---
		err := ctr.UnmarshalJSON(data)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "ID", ctr.ID)
		assert.Equal(t, "Image", ctr.Image)
		assert.Time(t, "2000-01-02T3:04:05Z", ctr.Created)
		assert.Empty(t, ctr.Env)
		assert.Empty(t, ctr.Labels)
	})

	t.Run("error - invalid time format", func(t *testing.T) {
		// --- Given ---
		data := []byte(`{
            "ID": "ID",
            "Created": "2006-01-02",
            "Config": { "Image": "Image" }
        }`)

		var ctr Container

		// --- When ---
		err := ctr.UnmarshalJSON(data)

		// --- Then ---
		assert.ErrorContain(t, "cannot parse", err)
		assert.Equal(t, "ID", ctr.ID)
		assert.Zero(t, ctr.Created)
	})

	t.Run("error - invalid JSON", func(t *testing.T) {
		// --- Given ---
		data := []byte(`{!!!}`)

		ctr := &Container{}

		// --- When ---
		err := ctr.UnmarshalJSON(data)

		// --- Then ---
		assert.ErrorContain(t, "invalid character", err)
	})
}

func Test_Containers_FindByImage(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		ctr0 := &Container{ID: "ID0", Image: "Image0"}
		ctr1 := &Container{ID: "ID1", Image: "Image1"}
		ctr2 := &Container{ID: "ID2", Image: "Image2"}

		cts := Containers{ctrs: []*Container{ctr0, ctr1, ctr2}}

		// --- When ---
		have, err := cts.FindByImage("Image3")

		// --- Then ---
		assert.NoError(t, err)
		assert.Nil(t, have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		cts := Containers{}

		// --- When ---
		have, err := cts.FindByImage("Image3")

		// --- Then ---
		assert.NoError(t, err)
		assert.Nil(t, have)
	})
}

func Test_Containers_FindByID(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		// --- Given ---
		ctr0 := &Container{ID: "ID0", Image: "Image0"}
		ctr1 := &Container{ID: "ID1", Image: "Image1"}
		ctr2 := &Container{ID: "ID2", Image: "Image2"}

		cts := Containers{ctrs: []*Container{ctr0, ctr1, ctr2}}

		// --- When ---
		have, err := cts.FindByID("ID3")

		// --- Then ---
		assert.NoError(t, err)
		assert.Nil(t, have)
	})

	t.Run("empty collection", func(t *testing.T) {
		// --- Given ---
		cts := Containers{}

		// --- When ---
		have, err := cts.FindByID("ID0")

		// --- Then ---
		assert.NoError(t, err)
		assert.Nil(t, have)
	})
}
