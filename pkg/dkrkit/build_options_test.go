// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"

	"github.com/ctx42/testkit/pkg/oskit"
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
		assert.Nil(t, have.bldRdr)
		assert.Equal(t, "", have.bldPth)
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
		assert.Nil(t, have.bldRdr)
		assert.Equal(t, "", have.bldPth)
		assert.Equal(t, "", have.iidPth)
		assert.False(t, have.noCache)
		assert.Nil(t, have.dryRun)
	})

	t.Run("error - WithBuildPth with WithBuildRdr", func(t *testing.T) {
		// --- Given ---
		pthOpt := WithBuildPth("Dockerfile")
		rdrOpt := WithBuildRdr(strings.NewReader("abc"))

		// --- When ---
		have, err := DefaultBuildOptions(pthOpt, rdrOpt)

		// --- Then ---
		wMsg := "WithBuildPth and WithBuildRdr are mutually exclusive"
		assert.ErrorEqual(t, wMsg, err)
		assert.Nil(t, have)
	})
}

func Test_BuildOptions_normalize(t *testing.T) {
	t.Run("generated values", func(t *testing.T) {
		// --- Given ---
		def := &BuildOptions{}
		env := []string{"KEY0=VAL0"}
		tmpDir := t.TempDir()

		// --- When ---
		have, cleanup := def.normalize(env, tmpDir)

		// --- Then ---
		assert.Equal(t, def.imgName+":"+def.imgTag, have)
		assert.Contain(t, "ctx42-tst-img-", def.imgName)
		assert.Contain(t, "ctx42-tst-tag-", def.imgTag)
		assert.False(t, def.noCache)

		assert.Equal(t, tmpDir, filepath.Dir(def.iidPth))
		assert.Contain(t, ".iid.log", def.iidPth)

		oskit.Create(t, "content", def.iidPth)
		cleanup()
		assert.NoFileExist(t, def.iidPth)
	})

	t.Run("set values are kept", func(t *testing.T) {
		// --- Given ---
		idfPth := oskit.Create(t, "content", t.TempDir(), "iid.log")
		def := &BuildOptions{
			imgName: "name",
			imgTag:  "tag",
			iidPth:  idfPth,
		}
		env := []string{"KEY0=VAL0"}
		tmpDir := t.TempDir()

		// --- When ---
		have, cleanup := def.normalize(env, tmpDir)

		// --- Then ---
		assert.Equal(t, "name:tag", have)
		assert.Equal(t, idfPth, def.iidPth)

		cleanup()
		assert.FileExist(t, idfPth)
	})

	t.Run("no cache set by env variable", func(t *testing.T) {
		// --- Given ---
		def := &BuildOptions{}
		env := []string{envBldNoCache + "=1"}
		tmpDir := t.TempDir()

		// --- When ---
		_, _ = def.normalize(env, tmpDir)

		// --- Then ---
		assert.True(t, def.noCache)
	})

	t.Run("no cache set by option", func(t *testing.T) {
		// --- Given ---
		def := &BuildOptions{noCache: true}
		env := []string{"KEY0=VAL0"}
		tmpDir := t.TempDir()

		// --- When ---
		_, _ = def.normalize(env, tmpDir)

		// --- Then ---
		assert.True(t, def.noCache)
	})
}

func Test_BuildOptions_cmdArgs(t *testing.T) {
	t.Run("all options", func(t *testing.T) {
		// --- Given ---
		def := &BuildOptions{
			iidPth:  "/tmp/iid.log",
			labels:  map[string]string{"lbl1": "v1", "lbl0": "v0"},
			args:    map[string]string{"ARG1": "v1", "ARG0": "v0"},
			noCache: true,
		}
		ref := "name:tag"
		env := []string{"SSH_AUTH_SOCK=/tmp/ssh.sock"}

		// --- When ---
		have := def.cmdArgs(ref, env)

		// --- Then ---
		want := []string{
			"build",
			"--rm",
			"-t", "name:tag",
			"--iidfile", "/tmp/iid.log",
			"--ssh=default",
			"--label", "lbl0=v0",
			"--label", "lbl1=v1",
			"--build-arg", "ARG0=v0",
			"--build-arg", "ARG1=v1",
			"--no-cache",
		}
		assert.Equal(t, want, have)
	})

	t.Run("no SSH_AUTH_SOCK", func(t *testing.T) {
		// --- Given ---
		def := &BuildOptions{iidPth: "/tmp/iid.log"}
		ref := "name:tag"
		env := []string{"KEY0=VAL0"}

		// --- When ---
		have := def.cmdArgs(ref, env)

		// --- Then ---
		want := []string{
			"build",
			"--rm",
			"-t", "name:tag",
			"--iidfile", "/tmp/iid.log",
		}
		assert.Equal(t, want, have)
	})
}

func Test_BuildOptions_source(t *testing.T) {
	t.Run("Dockerfile path", func(t *testing.T) {
		// --- Given ---
		def := &BuildOptions{bldPth: "/tmp/ctx/Containerfile"}

		// --- When ---
		hDir, hSin, hArgs := def.source()

		// --- Then ---
		assert.Equal(t, "/tmp/ctx", hDir)
		assert.Nil(t, hSin)
		assert.Equal(t, []string{"--file", "Containerfile", "."}, hArgs)
	})

	t.Run("Dockerfile path not set", func(t *testing.T) {
		// --- Given ---
		def := &BuildOptions{}

		// --- When ---
		hDir, hSin, hArgs := def.source()

		// --- Then ---
		assert.Equal(t, ".", hDir)
		assert.Nil(t, hSin)
		assert.Equal(t, []string{"--file", "Dockerfile", "."}, hArgs)
	})

	t.Run("Dockerfile reader", func(t *testing.T) {
		// --- Given ---
		rdr := strings.NewReader("FROM scratch")
		def := &BuildOptions{bldRdr: rdr}

		// --- When ---
		hDir, hSin, hArgs := def.source()

		// --- Then ---
		assert.Equal(t, "", hDir)
		assert.Same(t, rdr, hSin)
		assert.Equal(t, []string{"-"}, hArgs)
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

func Test_WithBuildRdr(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}
	rdr := strings.NewReader("abc")

	// --- When ---
	WithBuildRdr(rdr)(opts)

	// --- Then ---
	assert.Same(t, rdr, opts.bldRdr)
}

func Test_WithBuildPth(t *testing.T) {
	// --- Given ---
	opts := &BuildOptions{}

	// --- When ---
	WithBuildPth("/dir/path")(opts)

	// --- Then ---
	assert.Equal(t, "/dir/path", opts.bldPth)
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
