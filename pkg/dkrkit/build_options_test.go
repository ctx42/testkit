// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_DefaultBuildOptions(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		// --- When ---
		have, err := DefaultBuildOptions()

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "", have.imgName)
		assert.Equal(t, "", have.imgTag)
		assert.Nil(t, have.labels)
		assert.Nil(t, have.args)
		assert.Nil(t, have.dkfRdr)
		assert.Equal(t, "", have.dkfPth)
		assert.Equal(t, "", have.iidPth)
		assert.False(t, have.noCache)
		assert.Nil(t, have.dryRun)
	})

	t.Run("options applied", func(t *testing.T) {
		// --- Given ---
		opt := WithBuildName("name")

		// --- When ---
		have, err := DefaultBuildOptions(opt)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "name", have.imgName)
		assert.Equal(t, "", have.imgTag)
		assert.Nil(t, have.labels)
		assert.Nil(t, have.args)
		assert.Nil(t, have.dkfRdr)
		assert.Equal(t, "", have.dkfPth)
		assert.Equal(t, "", have.iidPth)
		assert.False(t, have.noCache)
		assert.Nil(t, have.dryRun)
	})

	t.Run("error - WithBuildDkfPth with WithBuildDkfRdr", func(t *testing.T) {
		// --- Given ---
		pthOpt := WithBuildDkfPth("testdata/simple/Dockerfile")
		rdrOpt := WithBuildDkfRdr(strings.NewReader("abc"))

		// --- When ---
		have, err := DefaultBuildOptions(pthOpt, rdrOpt)

		// --- Then ---
		wMsg := "WithBuildDkfPth and WithBuildDkfRdr are mutually exclusive"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})
}

func Test_WithBuildName(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	WithBuildName("name")(opts)

	// --- Then ---
	assert.Equal(t, "name", opts.imgName)
}

func Test_WithBuildTag(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	WithBuildTag("tag")(opts)

	// --- Then ---
	assert.Equal(t, "tag", opts.imgTag)
}

func Test_WithBuildLabel(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	WithBuildLabel("LAB0", "VAL0")(opts)

	// --- Then ---
	want := map[string]string{"LAB0": "VAL0"}
	assert.Equal(t, want, opts.labels)
}

func Test_WithBuildLabels(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		// --- Given ---
		opts := &BuildOptions{}
		labels := map[string]string{"A": "1", "B": "2"}

		// --- When ---
		WithBuildLabels(labels)(opts)

		// --- Then ---
		assert.Equal(t, map[string]string{"A": "1", "B": "2"}, opts.labels)
	})

	t.Run("passed map is copied", func(t *testing.T) {
		// --- Given ---
		opts := &BuildOptions{}
		labels := map[string]string{"A": "1", "B": "2"}

		// --- When ---
		WithBuildLabels(labels)(opts)

		// --- Then ---
		labels["B"] = "X"
		assert.Equal(t, map[string]string{"A": "1", "B": "X"}, labels)
		assert.Equal(t, map[string]string{"A": "1", "B": "2"}, opts.labels)
	})
}

func Test_WithBuildArg(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	WithBuildArg("A", "1")(opts)

	// --- Then ---
	assert.Equal(t, map[string]string{"A": "1"}, opts.args)
}

func Test_WithBuildArgs(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		// --- Given ---
		opts := &BuildOptions{}
		args := map[string]string{"A": "1", "B": "2"}

		// --- When ---
		WithBuildArgs(args)(opts)

		// --- Then ---
		assert.Equal(t, map[string]string{"A": "1", "B": "2"}, opts.args)
	})

	t.Run("passed map is copied", func(t *testing.T) {
		// --- Given ---
		opts := &BuildOptions{}
		args := map[string]string{"A": "1", "B": "2"}

		// --- When ---
		WithBuildArgs(args)(opts)

		// --- Then ---
		args["B"] = "X"
		assert.Equal(t, map[string]string{"A": "1", "B": "X"}, args)
		assert.Equal(t, map[string]string{"A": "1", "B": "2"}, opts.args)
	})
}

func Test_WithBuildDkfRdr(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}
	rdr := strings.NewReader("abc")

	// --- When ---
	WithBuildDkfRdr(rdr)(opts)

	// --- Then ---
	assert.Same(t, rdr, opts.dkfRdr)
}

func Test_WithBuildDkfPth(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	WithBuildDkfPth("/dir/path")(opts)

	// --- Then ---
	assert.Equal(t, "/dir/path", opts.dkfPth)
}

func Test_withBuildIIDFile(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	withBuildIIDFile("/tmp/iid.log")(opts)

	// --- Then ---
	assert.Equal(t, "/tmp/iid.log", opts.iidPth)
}

func Test_WithBuildNoCache(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	WithBuildNoCache()(opts)

	// --- Then ---
	assert.True(t, opts.noCache)
}

func Test_WithBuildDryRun(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}
	w := &bytes.Buffer{}

	// --- When ---
	WithBuildDryRun(w)(opts)

	// --- Then ---
	assert.Same(t, w, opts.dryRun)
}
