// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package selfkit_test

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ctx42/testkit/pkg/selfkit"
)

func ExampleNew() {
	// No selfkit flags: Run signals the caller to proceed with tests.
	se := selfkit.New(selfkit.WithArgs([]string{"prog"}))

	runTests, exitCode := se.Run(io.Discard, io.Discard)

	fmt.Println(runTests)
	fmt.Println(exitCode)
	// Output:
	// true
	// 0
}

func ExampleSelf_Run_toStdout() {
	var sout, eout strings.Builder
	args := []string{"prog", "--toStdout", "hello", "--toStderr", "world"}

	se := selfkit.New(selfkit.WithArgs(args))
	runTests, _ := se.Run(&sout, &eout)

	fmt.Println(sout.String())
	fmt.Println(eout.String())
	fmt.Println(runTests)
	// Output:
	// |sout: hello|
	// |eout: world|
	// false
}

func ExampleSelf_Run_exitCode() {
	args := []string{"prog", "--exitCode", "42"}

	se := selfkit.New(selfkit.WithArgs(args))
	runTests, exitCode := se.Run(io.Discard, io.Discard)

	fmt.Println(runTests)
	fmt.Println(exitCode)
	// Output:
	// false
	// 42
}

func ExampleSelf_Run_printEnv() {
	var sout strings.Builder
	_ = os.Setenv("MY_VAR", "secret")
	defer func() { _ = os.Unsetenv("MY_VAR") }()
	args := []string{"prog", "--printEnv", "MY_VAR"}

	se := selfkit.New(selfkit.WithArgs(args))
	runTests, _ := se.Run(&sout, io.Discard)

	fmt.Println(runTests)
	fmt.Println(sout.String())
	// Output:
	// false
	// |env: secret|
}

func ExampleSelf_Run_printArgs() {
	var sout strings.Builder
	args := []string{"prog", "--printArgs", "label", "arg1", "arg2"}

	se := selfkit.New(selfkit.WithArgs(args))
	_, _ = se.Run(&sout, io.Discard)

	fmt.Println(sout.String())
	// Output:
	// |args: label arg1,arg2|
}

func ExampleSelf_Run_noWrap() {
	var sout strings.Builder
	args := []string{"prog", "--noWrap", "--toStdout", "hello"}

	se := selfkit.New(selfkit.WithArgs(args))
	_, _ = se.Run(&sout, io.Discard)

	fmt.Println(sout.String())
	// Output:
	// hello
}

func ExampleSelf_Run_printToStderr() {
	_ = os.Setenv("MY_VAR", "value")
	defer func() { _ = os.Unsetenv("MY_VAR") }()
	var eout strings.Builder
	args := []string{"prog", "--printToStderr", "--printEnv", "MY_VAR"}

	se := selfkit.New(selfkit.WithArgs(args))
	_, _ = se.Run(io.Discard, &eout)

	fmt.Println(eout.String())
	// Output:
	// |env: value|
}
