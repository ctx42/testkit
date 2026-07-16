// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package modkit

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/testkit/pkg/oskit"
	"github.com/ctx42/testkit/pkg/pathkit"
	"github.com/ctx42/testkit/pkg/randkit"
)

func Test_Root(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- When ---
		have := Root()

		// --- Then ---
		want := must.Value(filepath.Abs("../.."))
		assert.Equal(t, want, have)
	})

	t.Run("could not find the project root", func(t *testing.T) {
		// --- Given ---
		dir := oskit.Chdir(t, pathkit.EvalSymlinks(t, t.TempDir()))

		// --- When ---
		msg := assert.PanicMsg(t, func() { Root() })

		// --- Then ---
		want := fmt.Sprintf("could not find go.mod starting at %s", dir)
		assert.Equal(t, want, *msg)
	})
}

func Test_Path(t *testing.T) {
	t.Run("root", func(t *testing.T) {
		// --- When ---
		have := Path()

		// --- Then ---
		want := must.Value(filepath.Abs("../.."))
		assert.Equal(t, want, have)
	})

	t.Run("relative", func(t *testing.T) {
		// --- When ---
		have := Path("build")

		// --- Then ---
		want := must.Value(filepath.Abs("../../build"))
		assert.Equal(t, want, have)
	})
}

func Test_Tmp(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		// Fake module in a temporary directory.
		dir := pathkit.EvalSymlinks(t, t.TempDir())
		oskit.Chdir(t, dir)
		oskit.Create(t, "", dir, "go.mod")
		sub := randkit.Str()

		// --- When ---
		have := Tmp(tspy, sub)

		// --- Then ---
		want := filepath.Join(Root(), "tmp", sub)
		assert.Equal(t, want, have)
		assert.DirExist(t, have)

		tspy.Finish().AssertExpectations()
		assert.NoDirExist(t, have)
	})

	t.Run("error - subs must not be empty slice", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogEqual("subdirectory name(s) must not be empty")
		tspy.Close()

		// --- When ---
		have := Tmp(tspy, "dir", "")

		// --- Then ---
		assert.Equal(t, "", have)
	})

	t.Run("error - sub must not be empty", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogEqual("the subdirectory name must not be empty")
		tspy.Close()

		// --- When ---
		have := Tmp(tspy, "")

		// --- Then ---
		assert.Equal(t, "", have)
	})
}

func Test_Ver(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- When ---
		have := Ver("github.com/ctx42/testing")

		// --- Then ---
		assert.Equal(t, "v0.55.0", have)
	})

	t.Run("failure", func(t *testing.T) {
		// --- Given ---
		dir := t.TempDir()
		oskit.Chdir(t, dir)

		// --- When ---
		msg := assert.PanicMsg(t, func() { Ver("github.com/ctx42/testing") })

		// --- Then ---
		wMsg := "could not find go.mod starting at %s"
		assert.Equal(t, fmt.Sprintf(wMsg, dir), *msg)
	})
}

func Test_ModVer(t *testing.T) {
	t.Run("error - invalid file", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/go.mod.invalid"

		// --- When ---
		ver, err := ModVer(pth, "github.com/dave/jennifer")

		// --- Then ---
		assert.ErrorContain(t, "no package", err)
		assert.Empty(t, ver)
	})

	t.Run("error - multiple candidates", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/go.mod.multiple"

		// --- When ---
		ver, err := ModVer(pth, "github.com/dave/jennifer")

		// --- Then ---
		assert.ErrorContain(t, "too many package", err)
		assert.Empty(t, ver)
	})

	t.Run("missing module", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/go.mod.example"

		// --- When ---
		ver, err := ModVer(pth, "github.com/rzajac/missing")

		// --- Then ---
		assert.ErrorContain(t, "no package", err)
		assert.Empty(t, ver)
	})

	t.Run("error - invalid file path", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/not_existing"

		// --- When ---
		ver, err := ModVer(pth, "github.com/dave/jennifer")

		// --- Then ---
		assert.ErrorIs(t, fs.ErrNotExist, err)
		assert.Empty(t, ver)
	})
}

func Test_ModVer_tabular(t *testing.T) {
	tt := []struct {
		testN string

		pkg string
		exp string
	}{
		{"1", "github.com/dave/jennifer", "v1.5.0"},
		{"2", "github.com/davecgh/go-spew", "v1.1.1"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			ver, err := ModVer("testdata/mod/go.mod.example", tc.pkg)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, ver)
		})
	}
}

func Test_GoVer(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/go.mod.example"

		// --- When ---
		ver, err := GoVer(pth)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "1.17", ver)
	})

	t.Run("error - invalid file", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/go.mod.no-go-version"

		// --- When ---
		ver, err := GoVer(pth)

		// --- Then ---
		assert.ErrorEqual(t, "invalid go.mod file", err)
		assert.Empty(t, ver)
	})

	t.Run("error - multiple candidates", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/go.mod.multiple-go-versions"

		// --- When ---
		ver, err := GoVer(pth)

		// --- Then ---
		assert.ErrorEqual(t, "invalid go.mod file", err)
		assert.Empty(t, ver)
	})

	t.Run("error - invalid file path", func(t *testing.T) {
		// --- Given ---
		pth := "testdata/mod/not_existing"

		// --- When ---
		ver, err := GoVer(pth)

		// --- Then ---
		assert.ErrorIs(t, fs.ErrNotExist, err)
		assert.Empty(t, ver)
	})
}
