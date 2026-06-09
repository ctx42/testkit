// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package randkit_test

import (
	"fmt"

	"github.com/ctx42/testkit/pkg/randkit"
)

func ExampleStr() {
	fmt.Println(randkit.Str(randkit.WithSeed(1)))
	// Output:
	// qLKZasgepC
}

func ExampleStr_withLen() {
	fmt.Println(randkit.Str(randkit.WithSeed(1), randkit.WithLen(6)))
	// Output:
	// qLKZas
}

func ExampleStr_withChars() {
	s := randkit.Str(
		randkit.WithChars(randkit.Digits),
		randkit.WithLen(8),
		randkit.WithSeed(1),
	)
	fmt.Println(s)
	// Output:
	// 37790310
}

func ExampleStr_withPrefixSuffix() {
	s := randkit.Str(
		randkit.WithPrefix("test-"),
		randkit.WithSuffix("-end"),
		randkit.WithLen(6),
		randkit.WithSeed(1),
	)
	fmt.Println(s)
	// Output:
	// test-qLKZas-end
}

func ExampleFileName() {
	fmt.Println(randkit.FileName("/tmp", randkit.WithSeed(1)))
	// Output:
	// /tmp/file-qLKZasg.txt
}

func ExampleFileName_withExt() {
	name := randkit.FileName("/tmp", randkit.WithExt(".json"), randkit.WithSeed(1))
	fmt.Println(name)
	// Output:
	// /tmp/file-qLKZasg.json
}

func ExampleInt() {
	fmt.Println(randkit.Int(100, randkit.WithSeed(1)))
	// Output:
	// 32
}

func ExamplePassword() {
	fmt.Println(randkit.Password(16, randkit.WithSeed(1)))
	// Output:
	// tSR9avhesITXkYun
}
