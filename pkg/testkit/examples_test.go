// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package testkit_test

import (
	"fmt"
	"strings"

	"github.com/ctx42/testkit/pkg/testkit"
)

func ExampleSHA1Reader() {
	r := strings.NewReader("hello")
	fmt.Println(testkit.SHA1Reader(r))
	// Output:
	// aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
}

func ExampleAddGlobalCleanup() {
	testkit.AddGlobalCleanup(func() {
		// cleanup logic, e.g. stopping a shared database container
	})
	testkit.RunGlobalCleanups()
	// Output:
}
