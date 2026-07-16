// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package selfkit

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_WithArgs(t *testing.T) {
	// --- Given ---
	args := []string{"arg0", "arg1"}
	slf := &Self{}

	// --- When ---
	WithArgs(args)(slf)

	// --- Then ---
	assert.Equal(t, []string{"arg0", "arg1"}, slf.args)
}

func Test_New(t *testing.T) {
	t.Run("empty args", func(t *testing.T) {
		// --- Given ---
		args := []string{}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.Equal(t, "", slf.toStdout)
		assert.Equal(t, "", slf.toStderr)
		assert.Equal(t, "", slf.printEnv)
		assert.Equal(t, "", slf.printArgs)
		assert.False(t, slf.printToStderr)
		assert.False(t, slf.noWrap)
		assert.Equal(t, -1, slf.exitCode)
	})

	t.Run("unknown arg", func(t *testing.T) {
		// --- Given ---
		args := []string{"program", "--unknown", "abc"}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.Equal(t, "", slf.toStdout)
		assert.Equal(t, "", slf.toStderr)
		assert.Equal(t, "", slf.printEnv)
		assert.Equal(t, "", slf.printArgs)
		assert.False(t, slf.printToStderr)
		assert.False(t, slf.noWrap)
		assert.Equal(t, -1, slf.exitCode)
	})

	t.Run("set toStdout", func(t *testing.T) {
		// --- Given ---
		args := []string{"program", "--toStdout", "abc"}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.Equal(t, "abc", slf.toStdout)
	})

	t.Run("set toStderr", func(t *testing.T) {
		// --- Given ---
		args := []string{"program", "--toStderr", "abc"}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.Equal(t, "abc", slf.toStderr)
	})

	t.Run("set printEnv", func(t *testing.T) {
		// --- Given ---
		args := []string{"program", "--printEnv", "abc"}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.Equal(t, "abc", slf.printEnv)
	})

	t.Run("set printToStderr", func(t *testing.T) {
		// --- Given ---
		args := []string{"program", "--printToStderr"}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.True(t, slf.printToStderr)
	})

	t.Run("set noWrap", func(t *testing.T) {
		// --- Given ---
		args := []string{"program", "--noWrap"}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.True(t, slf.noWrap)
	})

	t.Run("set printArgs", func(t *testing.T) {
		// --- Given ---
		args := []string{"program", "--printArgs", "abc"}

		// --- When ---
		slf := New(WithArgs(args))

		// --- Then ---
		assert.Equal(t, "abc", slf.printArgs)
	})
}

func Test_Self_Run(t *testing.T) {
	t.Run("no flags runs tests", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--unknown", "abc", "--other", "xyz", "-on"}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.True(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("print to stdout", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--toStdout", "abc"}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "|sout: abc|", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("print to stdout without wrap", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--noWrap", "--toStdout", " abc "}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, " abc ", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("print to stderr", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--toStderr", "abc"}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "", sout.String())
		assert.Equal(t, "|eout: abc|", eout.String())
	})

	t.Run("print to stderr without wrap", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--noWrap", "--toStderr", " abc "}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "", sout.String())
		assert.Equal(t, " abc ", eout.String())
	})

	t.Run("print env", func(t *testing.T) {
		// --- Given ---
		kv := fmt.Sprintf("kv%d", time.Now().UnixNano())
		t.Setenv(kv, kv)
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--printEnv", kv}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "|env: "+kv+"|", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("print env without wrap", func(t *testing.T) {
		// --- Given ---
		kv := fmt.Sprintf("kv%d", time.Now().UnixNano())
		t.Setenv(kv, kv)
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--noWrap", "--printEnv", kv}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, kv, sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("print env to stderr", func(t *testing.T) {
		// --- Given ---
		kv := fmt.Sprintf("kv%d", time.Now().UnixNano())
		t.Setenv(kv, kv)
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--printToStderr", "--printEnv", kv}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "", sout.String())
		assert.Equal(t, "|env: "+kv+"|", eout.String())
	})

	t.Run("print args", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--printArgs", "arg0", "arg1"}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "|args: arg0 arg1|", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("print args without wrap", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--noWrap", "--printArgs", " arg0", "arg1 "}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, " arg0 arg1 ", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("print args to stderr", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{
			"program",
			"--printToStderr",
			"--printArgs",
			"arg0",
			"arg1",
		}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "", sout.String())
		assert.Equal(t, "|args: arg0 arg1|", eout.String())
	})

	t.Run("set exit code", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--exitCode", "123"}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 123, exitCode)
		assert.Equal(t, "", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("set exit code zero", func(t *testing.T) {
		// --- Given ---
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{"program", "--exitCode", "0"}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, "", sout.String())
		assert.Equal(t, "", eout.String())
	})

	t.Run("do all", func(t *testing.T) {
		// --- Given ---
		kv := fmt.Sprintf("kv%d", time.Now().UnixNano())
		t.Setenv(kv, kv)
		sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
		args := []string{
			"program",
			"--printEnv", kv,
			"--toStdout", "abc",
			"--toStderr", "xyz",
			"--printArgs",
			"arg0", "arg1",
		}

		// --- When ---
		runTests, exitCode := New(WithArgs(args)).Run(sout, eout)

		// --- Then ---
		wantSout := "|sout: abc||env: " + kv + "||args: arg0 arg1|"
		assert.False(t, runTests)
		assert.Equal(t, 0, exitCode)
		assert.Equal(t, wantSout, sout.String())
		assert.Equal(t, "|eout: xyz|", eout.String())
	})
}
