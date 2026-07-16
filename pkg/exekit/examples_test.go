// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package exekit_test

import (
	"fmt"
	"testing"

	"github.com/ctx42/testkit/pkg/exekit"
)

func ExampleNew() {
	t := &testing.T{}
	exe := exekit.New(t, exekit.WithTrim)

	sout := exe.ExeStdout("echo", "hello")
	fmt.Println(sout)
	// Output:
	// hello
}

func ExampleIsWithCoverage() {
	args := []string{"-test.v", "-test.coverprofile=/tmp/cover.out"}
	path, ok := exekit.IsWithCoverage(args)
	fmt.Println(path)
	fmt.Println(ok)
	// Output:
	// /tmp/cover.out
	// true
}

func ExampleIsWithCoverage_absent() {
	args := []string{"-test.v", "-test.timeout=30s"}
	_, ok := exekit.IsWithCoverage(args)
	fmt.Println(ok)
	// Output:
	// false
}

func ExampleMaybeAddGoCovDir() {
	env := []string{"HOME=/root"}
	args := []string{"-test.coverprofile=/tmp/cover.out"}
	env = exekit.MaybeAddGoCovDir(env, args, func() string { return "/tmp/covdir" })
	fmt.Println(env)
	// Output:
	// [HOME=/root GOCOVERDIR=/tmp/covdir]
}
