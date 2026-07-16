// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package exekit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_IsWithCoverage_tabular(t *testing.T) {
	tt := []struct {
		testN string

		args     []string
		wantPath string
		wantIs   bool
	}{
		{
			"empty",
			[]string{},
			"",
			false,
		},
		{
			"coverprofile=",
			[]string{"a", "-coverprofile=/pth", "b"},
			"/pth",
			true,
		},
		{
			"test.coverprofile=",
			[]string{"a", "-test.coverprofile=/pth", "b"},
			"/pth",
			true,
		},
		{
			"coverprofile",
			[]string{"a", "-coverprofile", "/pth", "b"},
			"/pth",
			true,
		},
		{
			"test.coverprofile",
			[]string{"a", "-test.coverprofile", "/pth", "b"},
			"/pth",
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			havePath, haveIs := IsWithCoverage(tc.args)

			// --- Then ---
			assert.Equal(t, tc.wantPath, havePath)
			assert.Equal(t, tc.wantIs, haveIs)
		})
	}
}

func Test_MaybeAddGoCovDir(t *testing.T) {
	t.Run("not adding", func(t *testing.T) {
		// --- Given ---
		args := []string{"a", "b"}
		env := []string{"k0=v0"}

		// --- When ---
		have := MaybeAddGoCovDir(env, args, nil)

		// --- Then ---
		assert.Equal(t, []string{"k0=v0"}, have)
	})

	t.Run("adding", func(t *testing.T) {
		// --- Given ---
		args := []string{"a", "-test.coverprofile=/pth", "b"}
		env := []string{"k0=v0"}
		var tmp string
		getDir := func() string {
			tmp = t.TempDir()
			return tmp
		}

		// --- When ---
		have := MaybeAddGoCovDir(env, args, getDir)

		// --- Then ---
		assert.Equal(t, []string{"k0=v0", "GOCOVERDIR=" + tmp}, have)
	})

	t.Run("not overriding existing GOCOVERDIR", func(t *testing.T) {
		// --- Given ---
		args := []string{"a", "-test.coverprofile=/pth", "b"}
		env := []string{"k0=v0", "GOCOVERDIR=/existing"}

		// --- When ---
		have := MaybeAddGoCovDir(env, args, nil)

		// --- Then ---
		assert.Equal(t, []string{"k0=v0", "GOCOVERDIR=/existing"}, have)
	})
}
