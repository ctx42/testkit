// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package exekit

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/testkit/pkg/randkit"
)

func Test_WithWd(t *testing.T) {
	// --- Given ---
	ex := &Exe{}

	// --- When ---
	WithWd("/dir")(ex)

	// --- Then ---
	assert.Equal(t, "/dir", ex.wd)
}

func Test_WithEnvMap(t *testing.T) {
	// --- Given ---
	env := map[string]string{"k0": "v0"}
	ex := &Exe{}

	// --- When ---
	WithEnvMap(env)(ex)

	// --- Then ---
	assert.Equal(t, []string{"k0=v0"}, ex.env)
}

func Test_WithEnv(t *testing.T) {
	// --- Given ---
	env := []string{"k0=v0"}
	ex := &Exe{}

	// --- When ---
	WithEnv(env)(ex)

	// --- Then ---
	assert.Equal(t, []string{"k0=v0"}, ex.env)
}

func Test_WithTimeout(t *testing.T) {
	// --- Given ---
	ex := &Exe{}

	// --- When ---
	WithTimeout(time.Second)(ex)

	// --- Then ---
	assert.Equal(t, time.Second, ex.to)
}

func Test_WithExitCode(t *testing.T) {
	// --- Given ---
	ex := &Exe{}

	// --- When ---
	WithExitCode(123)(ex)

	// --- Then ---
	assert.Equal(t, 123, ex.ec)
}

func Test_WithLax(t *testing.T) {
	// --- Given ---
	ex := &Exe{}

	// --- When ---
	WithLax(ex)

	// --- Then ---
	assert.True(t, ex.lax)
}

func Test_WithDetCov(t *testing.T) {
	t.Run("coverage in args dir not set", func(t *testing.T) {
		// --- Given ---
		args := []string{"-coverprofile=/dir/cov.out"}
		ex := &Exe{
			t: t,
		}

		// --- When ---
		WithDetCov(args)(ex)

		// --- Then ---
		assert.Len(t, 1, ex.env)
		assert.True(t, strings.HasPrefix(ex.env[0], "GOCOVERDIR="))
	})

	t.Run("coverage in args dir set", func(t *testing.T) {
		// --- Given ---
		args := []string{"-coverprofile=/dir/cov.out"}
		ex := &Exe{
			env: []string{"GOCOVERDIR=/tmp/cov"},
		}

		// --- When ---
		WithDetCov(args)(ex)

		// --- Then ---
		assert.Len(t, 1, ex.env)
		assert.Equal(t, ex.env[0], "GOCOVERDIR=/tmp/cov")
	})

	t.Run("coverage not in args", func(t *testing.T) {
		// --- Given ---
		args := []string{"-other"}
		ex := &Exe{}

		// --- When ---
		WithDetCov(args)(ex)

		// --- Then ---
		assert.Empty(t, ex.env)
	})
}

func Test_Exe_Exe(t *testing.T) {
	t.Run("stdout intercepted", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--toStdout", "arg0")

		// --- Then ---
		assert.Equal(t, "|sout: arg0|", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("stdout trimmed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov, WithTrim)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--noWrap", "--toStdout", "arg0 ")

		// --- Then ---
		assert.Equal(t, "arg0", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("stderr intercepted", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--toStderr", "arg0")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "|eout: arg0|", eout)
	})

	t.Run("stderr not trimmed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--noWrap", "--toStderr", "arg0 ")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "arg0 ", eout)
	})

	t.Run("exit code 1", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogEqual("exit status 1")
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--exitCode", "1")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("exit code 2", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogEqual("exit status 2")
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--exitCode", "2")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("lax does not fail on non-zero exit code", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithLax, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--exitCode", "1")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("expect exit code 1", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithExitCode(1), WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--exitCode", "1")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("expect exit code 1 got 0", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("expected exit code 1 got 0")
		tspy.Close()

		exe := New(tspy, WithExitCode(1), WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--toStdout", "ok")

		// --- Then ---
		assert.Equal(t, "|sout: ok|", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("print not existing env var", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--printEnv", randkit.Str())

		// --- Then ---
		assert.Equal(t, "|env: |", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("print existing env var", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		kv := randkit.Str()
		t.Setenv(kv, kv)

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--printEnv", kv)

		// --- Then ---
		assert.Equal(t, "|env: "+kv+"|", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("with custom map env", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		kv := randkit.Str()
		env := envSplit(os.Environ())
		env[kv] = kv

		exe := New(tspy, WithEnvMap(env), WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--printEnv", kv)

		// --- Then ---
		assert.Equal(t, "|env: "+kv+"|", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("with custom os env", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		kv := randkit.Str()
		env := append(os.Environ(), kv+"="+kv)

		exe := New(tspy, WithEnv(env), WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--printEnv", kv)

		// --- Then ---
		assert.Equal(t, "|env: "+kv+"|", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("with custom working directory", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithWd("testdata/dir/sub"), WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe("ls", "-1")

		// --- Then ---
		assert.Equal(t, "file.txt\nsub_sub\n", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("set execution timeout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogEqual("command timed out after 500ms: /usr/bin/sleep 1")
		tspy.Close()

		exe := New(tspy, WithTimeout(500*time.Millisecond), WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe("sleep", "1")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("set execution timeout ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithTimeout(2*time.Second), WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe("sleep", "1")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "", eout)
	})

	t.Run("expect exit code 1 trim", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithExitCode(1), WithTrim, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(
			os.Args[0],
			"--noWrap", "--toStdout", "arg0 ",
			"--toStderr", " err ", "--exitCode", "1",
		)

		// --- Then ---
		assert.Equal(t, "arg0", sout)
		assert.Equal(t, "err", eout)
	})

	t.Run("fail with stderr", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("|eout: msg|")
		tspy.ExpectLogContain("exit status 1")
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout, eout := exe.Exe(os.Args[0], "--toStderr", "msg", "--exitCode", "1")

		// --- Then ---
		assert.Equal(t, "", sout)
		assert.Equal(t, "|eout: msg|", eout)
	})
}

func Test_Exe_Stdout(t *testing.T) {
	t.Run("stdout intercepted", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout := exe.ExeStdout(os.Args[0], "--toStdout", "arg0")

		// --- Then ---
		assert.Equal(t, "|sout: arg0|", sout)
	})

	t.Run("stdout not trimmed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout := exe.ExeStdout(os.Args[0], "--noWrap", "--toStdout", "arg0 ")

		// --- Then ---
		assert.Equal(t, "arg0 ", sout)
	})

	t.Run("error if stderr written to", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "expected empty stderr got: \"|eout: arg1|\""
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		sout := exe.ExeStdout(os.Args[0], "--toStdout", "arg0", "--toStderr", "arg1")

		// --- Then ---
		assert.Equal(t, "|sout: arg0|", sout)
	})
}

func Test_Exe_Stderr(t *testing.T) {
	t.Run("stderr intercepted", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		eout := exe.ExeStderr(os.Args[0], "--toStderr", "arg0")

		// --- Then ---
		assert.Equal(t, "|eout: arg0|", eout)
	})

	t.Run("stderr not trimmed", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		eout := exe.ExeStderr(os.Args[0], "--noWrap", "--toStderr", "arg0 ")

		// --- Then ---
		assert.Equal(t, "arg0 ", eout)
	})

	t.Run("error if stderr written to", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "expected empty stdout got: \"|sout: arg0|\""
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		exe := New(tspy, WithDevOsCov)

		// --- When ---
		eout := exe.ExeStderr(os.Args[0], "--toStdout", "arg0", "--toStderr", "arg1")

		// --- Then ---
		assert.Equal(t, "|eout: arg1|", eout)
	})
}
