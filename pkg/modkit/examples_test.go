// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package modkit_test

import (
	"fmt"
	"path/filepath"

	"github.com/ctx42/testkit/pkg/modkit"
)

func ExampleRoot() {
	// Root walks up from the working directory to the module root (the
	// directory that holds go.mod).
	fmt.Println(filepath.Base(modkit.Root()))
	// Output:
	// testkit
}

func ExamplePath() {
	// Path joins its segments onto the module root; the target need not
	// exist.
	fmt.Println(filepath.Base(modkit.Path("build")))
	// Output:
	// build
}

func ExampleModVer() {
	// ModVer reports the version a go.mod pins for a given module.
	ver, _ := modkit.ModVer(
		"testdata/mod/go.mod.example",
		"github.com/dave/jennifer",
	)
	fmt.Println(ver)
	// Output:
	// v1.5.0
}

func ExampleGoVer() {
	// GoVer reports the Go version declared in a go.mod file.
	ver, _ := modkit.GoVer("testdata/mod/go.mod.example")
	fmt.Println(ver)
	// Output:
	// 1.17
}
