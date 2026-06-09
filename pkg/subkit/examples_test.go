// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package subkit_test

import (
	"fmt"

	"github.com/ctx42/testkit/pkg/subkit"
)

func ExampleNew() {
	sub := subkit.New("Test_MyFunc/success")
	if sub.InMainProcess() {
		// sout, eout, err := sub.Run()
		return
	}
	// --- IN SUBPROCESS ---
	// The real test body runs here in the child process.

	// Output:
}

func ExampleNewPkg() {
	sub := subkit.NewPkg("github.com/ctx42/testkit/pkg/myservice")
	// sout, eout, err := sub.Run()
	_ = sub
	// Output:
}

func ExampleSubProcess_InSubProcess() {
	sub := subkit.New("Test_Example")
	fmt.Println(sub.InSubProcess())
	// Output:
	// false
}

func ExampleSubProcess_InMainProcess() {
	sub := subkit.New("Test_Example")
	fmt.Println(sub.InMainProcess())
	// Output:
	// true
}

func ExampleGetCovProfile() {
	args := []string{"-test.v", "-test.coverprofile", "/tmp/cover.out"}
	fmt.Println(subkit.GetCovProfile(args))
	// Output:
	// [-test.coverprofile /tmp/cover.out]
}

func ExampleGetCovProfile_notPresent() {
	args := []string{"-test.v", "-test.timeout", "30s"}
	fmt.Println(subkit.GetCovProfile(args))
	// Output:
	// []
}
