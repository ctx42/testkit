package exekit

import "fmt"

func ExampleIsWithCoverage() {
	args := []string{"-test.v", "-test.coverprofile=/tmp/cover.out"}
	path, ok := IsWithCoverage(args)
	fmt.Println(path)
	fmt.Println(ok)
	// Output:
	// /tmp/cover.out
	// true
}

func ExampleIsWithCoverage_absent() {
	args := []string{"-test.v", "-test.timeout=30s"}
	_, ok := IsWithCoverage(args)
	fmt.Println(ok)
	// Output:
	// false
}

func ExampleMaybeAddGoCovDir() {
	env := []string{"HOME=/root"}
	args := []string{"-test.coverprofile=/tmp/cover.out"}
	env = MaybeAddGoCovDir(env, args, func() string { return "/tmp/covdir" })
	fmt.Println(env)
	// Output:
	// [HOME=/root GOCOVERDIR=/tmp/covdir]
}
