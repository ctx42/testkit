// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit_test

import (
	"fmt"
	"strings"

	"github.com/ctx42/testkit/pkg/dkrkit"
)

func ExampleRef() {
	// Ref assembles an image reference from repo, name, and tag.
	ref := dkrkit.Ref("example.com/repo", "myapp", "v1.2.3")
	fmt.Println(ref)
	// Output:
	// example.com/repo/myapp:v1.2.3
}

func ExampleRef_noRepo() {
	// Without a repo, Ref returns name:tag.
	ref := dkrkit.Ref("", "myapp", "v1.2.3")
	fmt.Println(ref)
	// Output:
	// myapp:v1.2.3
}

func ExampleStripHashName() {
	// StripHashName removes the algorithm prefix from a digest.
	id := dkrkit.StripHashName("sha256:b3aab1576e98b7f41f01fa")
	fmt.Println(id)
	// Output:
	// b3aab1576e98b7f41f01fa
}

func ExampleStripHashName_noPrefix() {
	// A bare hex ID is returned unchanged.
	id := dkrkit.StripHashName("b3aab1576e98b7f41f01fa")
	fmt.Println(id)
	// Output:
	// b3aab1576e98b7f41f01fa
}

func ExampleShortID() {
	// ShortID truncates a long hex ID to its first 12 characters.
	short := dkrkit.ShortID("785e9f61d4598b65c6c86c5f122830f7")
	fmt.Println(short)
	// Output:
	// 785e9f61d459
}

func ExampleShortID_nonHex() {
	// Non-hex strings (e.g. image references) are returned unchanged.
	ref := dkrkit.ShortID("myapp:v1.2.3")
	fmt.Println(ref)
	// Output:
	// myapp:v1.2.3
}

func ExampleToBuildArgs() {
	// ToBuildArgs converts a string map to the pointer map that Docker
	// build arguments require.
	args := dkrkit.ToBuildArgs(map[string]string{"VERSION": "v1.2.3"})
	fmt.Println(*args["VERSION"])
	// Output:
	// v1.2.3
}

func ExampleRandName() {
	// RandName returns a unique name suitable for a test image.
	name := dkrkit.RandName()
	fmt.Println(strings.HasPrefix(name, "ctx42-tst-img-"))
	// Output:
	// true
}

func ExampleRandTag() {
	// RandTag returns a unique tag suitable for a test image.
	tag := dkrkit.RandTag()
	fmt.Println(strings.HasPrefix(tag, "ctx42-tst-tag-"))
	// Output:
	// true
}

func ExampleRandRef() {
	// RandRef returns a unique name:tag reference for a test image.
	ref := dkrkit.RandRef()
	parts := strings.SplitN(ref, ":", 2)
	fmt.Println(strings.HasPrefix(parts[0], "ctx42-tst-img-"))
	fmt.Println(strings.HasPrefix(parts[1], "ctx42-tst-tag-"))
	// Output:
	// true
	// true
}

func ExampleRandNet() {
	// RandNet returns a unique name suitable for a test Docker network.
	name := dkrkit.RandNet()
	fmt.Println(strings.HasPrefix(name, "ctx42-tst-net-"))
	// Output:
	// true
}
