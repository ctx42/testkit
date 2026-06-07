// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package exekit provides test helpers for running external commands and
// asserting their exit code, standard output, and standard error.
//
// The central type is [Exe], which wraps [os/exec] with a configurable
// timeout, environment, and exit-code expectation. Construct one with [New]
// and run commands with [Exe.Exe], [Exe.ExeStdout], or [Exe.ExeStderr].
//
// Coverage helpers ([WithDetCov], [WithDevOsCov], [MaybeAddGoCovDir],
// [IsWithCoverage]) assist with propagating Go coverage instrumentation
// into sub-processes launched during tests.
package exekit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ctx42/testing/pkg/tester"
)

// DefaultTimeout is the timeout applied to every new [Exe] instance created
// by [New]. Prefer [WithTimeout] to override it for a single instance.
//
// WARNING: this is package-level mutable state shared across all goroutines.
// Mutating it in one test affects all concurrent and subsequent [New] calls.
var DefaultTimeout = 5 * time.Second

// WithWd sets [Exe] working directory for the command.
func WithWd(wd string) func(*Exe) { return func(ex *Exe) { ex.wd = wd } }

// WithEnvMap sets [Exe] environment for a command.
func WithEnvMap(env map[string]string) func(*Exe) {
	return func(ex *Exe) { ex.env = envJoin(env) }
}

// WithEnv sets [Exe] environment for a command.
func WithEnv(env []string) func(*Exe) {
	return func(ex *Exe) { ex.env = env }
}

// WithTimeout sets [Exe] timeout for a command execution.
func WithTimeout(to time.Duration) func(*Exe) {
	return func(ex *Exe) { ex.to = to }
}

// WithExitCode sets expected Exe exit code for a command execution.
func WithExitCode(ec int) func(*Exe) {
	return func(ex *Exe) { ex.ec = ec }
}

// WithDetCov is a special option detecting if the current test binary was
// compiled with coverage, and if it was, it adds to the [Exe] environment the
// GOCOVERDIR variable set to test case generated temporary directory. If
// the GOCOVERDIR variable already exists in the [Exe] environment, it will not
// be overridden.
//
// Make sure this option is the last on the list of options to [Exe].
func WithDetCov(args []string) func(*Exe) {
	return func(ex *Exe) {
		if _, ok := IsWithCoverage(args); !ok {
			return
		}
		for _, kv := range ex.env {
			if strings.HasPrefix(kv, "GOCOVERDIR") {
				return
			}
		}
		ex.env = append(ex.env, "GOCOVERDIR="+os.TempDir())
	}
}

// WithDevOsCov is a convenience wrapper around [WithDetCov] that uses
// [os.Args]. Apply it last in the option list so it sees the final
// environment.
func WithDevOsCov(ex *Exe) { WithDetCov(os.Args)(ex) }

// WithTrim makes [Exe] to trim whitespace from stdout and stderr using
// [strings.TrimSpace] before returning it.
func WithTrim(ex *Exe) { ex.trim = true }

// Exe represents command executor.
type Exe struct {
	// Specifies the process's standard input.
	sin io.Reader

	// Specifies the working directory of the command.
	wd string

	// Specifies the environment of the process.
	// If not set, os.Environ will be used.
	env []string

	// Execution timeout. By default, it is set to 5 seconds.
	to time.Duration

	// Expected exit code.
	ec int

	// Trim output.
	trim bool

	// Test manager.
	t tester.T
}

// New returns new instance of Exe.
// By default, the environment for the command to run will be set using
// [os.Environ], if you wish to change it use [WithEnv].
func New(t tester.T, opts ...func(*Exe)) *Exe {
	ex := &Exe{
		sin: &bytes.Buffer{},
		env: os.Environ(),
		to:  DefaultTimeout,
		t:   t,
	}
	for _, opt := range opts {
		opt(ex)
	}
	return ex
}

// Exe runs cmd with the given args and returns what the command wrote to stdout
// and stderr. On error, it marks the test as failed. It always returns both
// outputs.
func (ex *Exe) Exe(cmd string, args ...string) (string, string) {
	ex.t.Helper()
	ctx := context.Background()
	if ex.to > 0 {
		var cxl func()
		ctx, cxl = context.WithTimeout(ctx, ex.to)
		defer cxl()
	}

	sout, eout := &bytes.Buffer{}, &bytes.Buffer{}
	c := exec.CommandContext(ctx, cmd, args...)
	c.Env = ex.env
	c.Stdout = sout
	c.Stderr = eout
	c.Stdin = ex.sin
	c.Dir = ex.wd

	if err := c.Run(); err != nil {
		so, eo := sout.String(), eout.String()
		if ex.ec != 0 {
			ee, ok := errors.AsType[*exec.ExitError](err)
			if ok && ex.ec == ee.ExitCode() {
				if ex.trim {
					so = strings.TrimSpace(so)
					eo = strings.TrimSpace(eo)
				}
				return so, eo
			}
		}
		if eo != "" {
			ex.t.Log(strings.TrimSpace(eo))
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			ex.t.Errorf("command timed out after %s: %s", ex.to, c)
		} else {
			ex.t.Error(err)
		}
	}

	if ex.ec != 0 {
		ex.t.Errorf("expected exit code %d got 0", ex.ec)
	}

	so, eo := sout.String(), eout.String()
	if ex.trim {
		so = strings.TrimSpace(so)
		eo = strings.TrimSpace(eo)
	}
	return so, eo
}

// ExeStdout runs cmd with the given args and returns the command's stdout. On
// error, it marks the test as failed. It always returns what was written to
// stdout.
func (ex *Exe) ExeStdout(cmd string, args ...string) string {
	ex.t.Helper()
	sout, eout := ex.Exe(cmd, args...)
	if eout != "" {
		ex.t.Errorf("expected empty stderr got: %#v", eout)
	}
	return sout
}

// ExeStderr runs cmd with the given args and returns the command's stderr. On
// error, it marks the test as failed. It always returns what was written to
// stderr.
func (ex *Exe) ExeStderr(cmd string, args ...string) string {
	ex.t.Helper()
	sout, eout := ex.Exe(cmd, args...)
	if sout != "" {
		ex.t.Errorf("expected empty stdout got: %#v", sout)
	}
	return eout
}
