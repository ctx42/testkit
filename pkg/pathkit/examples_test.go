// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package pathkit_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/ctx42/testkit/pkg/pathkit"
)

func ExampleAbsPath() {
	t := &testing.T{}

	// AbsPath joins the segments, then resolves them to an absolute path.
	abs := pathkit.AbsPath(t, ".", "testdata")

	fmt.Println(filepath.IsAbs(abs))
	fmt.Println(filepath.Base(abs))
	// Output:
	// true
	// testdata
}

func ExampleEvalSymlinks() {
	t := &testing.T{}

	// testdata/dir_sym_link is a symlink to testdata/dir; EvalSymlinks
	// resolves it to the real path.
	resolved := pathkit.EvalSymlinks(t, "testdata", "dir_sym_link")

	fmt.Println(resolved)
	// Output:
	// testdata/dir
}
