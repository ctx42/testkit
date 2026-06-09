// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package subkit

import (
	"fmt"
	"os"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

// Constants used during subProcess testing.
const (
	// testSubprocessFail represents an environment variable which value is
	// used during subProcess instance testing. When set to "1" the subprocess
	// test should fail, if not set or set to "0" subprocess should succeed.
	TestSubProcFail = "TEST_SUBPROCESS_FAIL"

	// TestSubProcValue represents an environment variable which value is
	// used during subProcess instance testing. Its value is always printed by
	//  a test running in a subprocess.
	TestSubProcValue = "TEST_SUBPROCESS_VALUE"
)

func Test_SubProcess(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		sub := New(t.Name())
		if sub.InMainProcess() {
			t.Setenv(TestSubProcFail, "0")
			t.Setenv(TestSubProcValue, "123")
			// --- TEST SUBPROCESS OUTPUT AND ERROR ---
			sout, eout, err := sub.Run()
			assert.NoError(t, err)
			wMsg := fmt.Sprintf("%s: %s", TestSubProcValue, "123")
			assert.Contain(t, wMsg, sout)
			assert.Empty(t, eout)
			return
		}
		// --- IN SUBPROCESS ---

		// --- Then ---
		t.Logf("%s: %s", TestSubProcValue, os.Getenv(TestSubProcValue))
		if val := os.Getenv(TestSubProcFail); val == "1" {
			t.Error("subprocess test asked to fail")
		}
	})

	t.Run("failure", func(t *testing.T) {
		// --- Given ---
		sub := New(t.Name())
		if sub.InMainProcess() {
			t.Setenv(TestSubProcFail, "1")
			t.Setenv(TestSubProcValue, "123")
			// --- TEST SUBPROCESS OUTPUT AND ERROR ---
			sout, eout, err := sub.Run()
			assert.ExitCode(t, 1, err)
			wMsg := fmt.Sprintf("%s: %s", TestSubProcValue, "123")
			assert.Contain(t, wMsg, sout)
			assert.Empty(t, eout)
			return
		}
		// --- IN SUBPROCESS ---

		// --- Then ---
		t.Logf("%s: %s", TestSubProcValue, os.Getenv(TestSubProcValue))
		if val := os.Getenv(TestSubProcFail); val == "1" {
			t.Error("subprocess test asked to fail")
		}
	})
}

func Test_subProcess_InSubProcess_InMainProcess(t *testing.T) {
	t.Run("in subprocess", func(t *testing.T) {
		// --- Given ---
		name := "TEST_NAME_IN"
		t.Setenv(envNamePrefix+name, "1")
		sub := New(name)

		// --- When ---
		haveSub := sub.InSubProcess()
		haveMain := sub.InMainProcess()

		// --- Then ---
		assert.True(t, haveSub)
		assert.False(t, haveMain)
	})

	t.Run("not in subprocess", func(t *testing.T) {
		// --- Given ---
		name := "TEST_NAME_NOT_IN"
		sub := New(name)

		// --- When ---
		haveSub := sub.InSubProcess()
		haveMain := sub.InMainProcess()

		// --- Then ---
		assert.False(t, haveSub)
		assert.True(t, haveMain)
	})
}

func Test_NewPkg(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- When ---
		sub := NewPkg("github.com/ctx42/testkit/pkg/subkit")

		// --- Then ---
		assert.Equal(t, "github.com/ctx42/testkit/pkg/subkit", sub.name)
		assert.True(t, sub.pkg)
	})
}

func Test_subProcess_envName(t *testing.T) {
	t.Run("test name", func(t *testing.T) {
		// --- Given ---
		name := "TEST_NAME_NOT_IN"
		sub := New(name)

		// --- When ---
		have := sub.envName()

		// --- Then ---
		assert.Equal(t, envNamePrefix+name, have)
	})

	t.Run("pkg path slashes replaced", func(t *testing.T) {
		// --- Given ---
		sub := NewPkg("github.com/ctx42/testkit/pkg/subkit")

		// --- When ---
		have := sub.envName()

		// --- Then ---
		want := envNamePrefix + "github.com_ctx42_testkit_pkg_subkit"
		assert.Equal(t, want, have)
	})
}

func Test_GetCovProfile(t *testing.T) {
	t.Run("main run with coverage", func(t *testing.T) {
		// --- Given ---
		args := []string{"-test.coverprofile", "/path"}

		// --- When ---
		have := GetCovProfile(args)

		// --- Then ---
		assert.Equal(t, args, have)
	})

	t.Run("coverage arg with no path", func(t *testing.T) {
		// --- Given ---
		args := []string{"-test.coverprofile"}

		// --- When ---
		have := GetCovProfile(args)

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("main run without coverage", func(t *testing.T) {
		// --- Given ---
		args := []string{"-test.v", "/path"}

		// --- When ---
		have := GetCovProfile(args)

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("args nil", func(t *testing.T) {
		// --- When ---
		have := GetCovProfile(nil)

		// --- Then ---
		assert.Nil(t, have)
	})
}
