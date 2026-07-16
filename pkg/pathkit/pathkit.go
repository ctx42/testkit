// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package pathkit provides test helpers for common [path/filepath]
// operations. Every exported function integrates with [tester.T]: on
// error it marks the test as failed, writes a diagnostic to the test
// log, and returns a safe zero value so the test can continue
// executing.
package pathkit

import (
	"path/filepath"

	"github.com/ctx42/testing/pkg/tester"
)

// AbsPath is a wrapper around [filepath.Abs]. The path is constructed from pth
// and elems like the [filepath.Join] function. On error, it marks the test as
// failed and returns an empty string.
func AbsPath(t tester.T, pth string, elems ...string) string {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	ret, err := filepath.Abs(pth)
	if err != nil {
		t.Error(err)
	}
	return ret
}

// EvalSymlinks is a wrapper around [filepath.EvalSymlinks]. The path is
// constructed from pth and elems like the [filepath.Join] function. On error,
// it marks the test as failed and returns an empty string.
func EvalSymlinks(t tester.T, pth string, elems ...string) string {
	t.Helper()
	pth = filepath.Join(append([]string{pth}, elems...)...)
	ret, err := filepath.EvalSymlinks(pth)
	if err != nil {
		t.Error(err)
	}
	return ret
}
