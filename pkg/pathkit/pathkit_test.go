// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package pathkit

import (
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_AbsPath(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := AbsPath(tspy, ".")

		// --- Then ---
		want := must.Value(filepath.Abs("."))
		assert.Equal(t, want, have)
	})

	t.Run("join path", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		want := must.Value(filepath.Abs("."))

		// --- When ---
		have := AbsPath(tspy, ".", "dir", "abc.txt")

		// --- Then ---
		want = filepath.Join(want, "dir", "abc.txt")
		assert.Equal(t, want, have)
	})
}

func Test_EvalSymlinks(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := EvalSymlinks(tspy, "testdata/dir_sym_link")

		// --- Then ---
		assert.Equal(t, "testdata/dir", have)
	})

	t.Run("path joined", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := EvalSymlinks(tspy, "testdata", "dir_sym_link")

		// --- Then ---
		assert.Equal(t, "testdata/dir", have)
	})

	t.Run("error - not existing file", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("no such file or directory")
		tspy.ExpectLogContain("testdata/not_existing")
		tspy.Close()

		// --- When ---
		have := EvalSymlinks(tspy, "testdata/not_existing")

		// --- Then ---
		assert.Equal(t, "", have)
	})
}
