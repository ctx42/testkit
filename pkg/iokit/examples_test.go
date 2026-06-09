// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package iokit_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
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

func ExampleErrReader() {
	rdr := strings.NewReader("some text")
	rcs := iokit.ErrReader(rdr, 3)

	data, err := io.ReadAll(rcs)

	fmt.Printf("error: %v\n", err)
	fmt.Printf(" data: %s\n", string(data))
	// Output:
	// error: read error
	//  data: som
}

func ExampleErrReader_custom_error() {
	mye := errors.New("my error")
	rdr := strings.NewReader("some text")
	rcs := iokit.ErrReader(rdr, 4, iokit.WithReadErr(mye))

	data, err := io.ReadAll(rcs)

	fmt.Printf("error: %v\n", err)
	fmt.Printf(" data: %s\n", string(data))
	// Output:
	// error: my error
	//  data: some
}

func ExampleErrWriter() {
	dst := &bytes.Buffer{}
	ce := errors.New("my error")
	ew := iokit.ErrWriter(dst, 3, iokit.WithWriteErr(ce))

	n, err := ew.Write([]byte{0, 1, 2, 3})

	fmt.Printf("    n: %d\n", n)
	fmt.Printf("error: %v\n", err)
	fmt.Printf("  dst: %v\n", dst.Bytes())
	// Output:
	//     n: 3
	// error: my error
	//   dst: [0 1 2]
}
