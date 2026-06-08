// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package iokit_test

import (
	"fmt"
	"testing"

	"github.com/ctx42/testkit/pkg/iokit"
)

func ExampleWetBuffer() {
	t := &testing.T{}

	buf := iokit.WetBuffer(t, "stdout")
	_, _ = buf.WriteString("hello")

	fmt.Println(buf.String())
	// Output:
	// hello
}

func ExampleDryBuffer() {
	t := &testing.T{}

	errOut := iokit.DryBuffer(t, "stderr")

	// Pass errOut as io.Writer to code that should produce no output.
	// Cleanup calls t.Error if anything is written.
	_ = errOut
	// Output:
}

func ExampleBuffer_SkipExamine() {
	t := &testing.T{}

	buf := iokit.WetBuffer(t, "stdout").SkipExamine()
	_, _ = buf.WriteString("data")
	// buf.String() need not be called — SkipExamine disables the check.

	fmt.Println(buf.Kind())
	// Output:
	// wet
}
