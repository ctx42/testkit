// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_withCmdWD(t *testing.T) {
	// --- Given ---
	opts := &cmdOptions{}

	// --- When ---
	withCmdWD("/dir/path")(opts)

	// --- Then ---
	assert.Equal(t, "/dir/path", opts.wd)
}

func Test_withCmdStdin(t *testing.T) {
	// --- Given ---
	r := strings.NewReader("abc")
	opts := &cmdOptions{}

	// --- When ---
	withCmdStdin(r)(opts)

	// --- Then ---
	assert.Same(t, r, opts.sin)
}

func Test_dockerCmd(t *testing.T) {
	t.Run("cancel context", func(t *testing.T) {
		// --- Given ---
		// Force-remove any containers still using the image before the image
		// cleanup registered by BuildTestImg runs (LIFO cleanup order).
		t.Cleanup(func() {
			ctx := context.Background()
			env := os.Environ()
			args := []string{
				"container",
				"ls",
				"-q",
				"--filter", "ancestor=" + TestImg0.iid,
			}
			if cid, _, _ := dockerCmd(ctx, env, args); cid != "" {
				args = []string{"container", "rm", "-f", cid}
				_, _, _ = dockerCmd(ctx, env, args)
			}
		})

		ctx, cxl := context.WithTimeout(context.Background(), time.Second)
		defer cxl()
		env := os.Environ()

		args := []string{"run", "--rm", TestImg0.iid, "sleep", "100"}

		// --- When ---
		haveSout, haveEout, err := dockerCmd(ctx, env, args)

		// --- Then ---
		wMsg := "" +
			"docker command error:\n" +
			"  cmd: docker run --rm %s sleep 100\n" +
			"  err: signal: killed"
		wMsg = fmt.Sprintf(wMsg, TestImg0.iid)
		assert.ErrorEqual(t, wMsg, err)
		assert.Empty(t, haveSout)
		assert.Empty(t, haveEout)
	})

	t.Run("trims standard output", func(t *testing.T) {
		// --- Given ---
		ctx := t.Context()
		env := os.Environ()
		args := []string{"run", "--rm", TestImg0.iid, "echo", "abc\n"}

		// --- When ---
		haveSout, haveEout, err := dockerCmd(ctx, env, args)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, "abc", haveSout)
		assert.Empty(t, haveEout)
	})

	t.Run("trims standard error", func(t *testing.T) {
		// --- Given ---
		ctx := t.Context()
		env := os.Environ()
		args := []string{"not_existing"}

		// --- When ---
		haveSout, haveEout, err := dockerCmd(ctx, env, args)

		// --- Then ---
		assert.ExitCode(t, 1, err)
		assert.Empty(t, haveSout)
		assert.False(t, strings.HasSuffix(haveEout, "\n"))
	})
}
