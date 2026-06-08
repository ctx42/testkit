package testkit

import (
	"fmt"
	"strings"
)

func ExampleSHA1Reader() {
	r := strings.NewReader("hello")
	fmt.Println(SHA1Reader(r))
	// Output:
	// aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
}

func ExampleAddGlobalCleanup() {
	AddGlobalCleanup(func() {
		// cleanup logic, e.g. stopping a shared database container
	})
	RunGlobalCleanups()
	// Output:
}
