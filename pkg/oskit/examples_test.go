// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package oskit_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ctx42/testkit/pkg/oskit"
)

func ExampleReadFile() {
	t := &testing.T{}
	data := oskit.ReadFile(t, "testdata/file.txt")
	fmt.Println(string(data))
	// Output:
	// content
}

func ExampleReadFileStr() {
	t := &testing.T{}
	s := oskit.ReadFileStr(t, "testdata/file.txt")
	fmt.Println(s)
	// Output:
	// content
}

func ExampleWriteStr() {
	t := &testing.T{}
	dir, _ := os.MkdirTemp("", "oskit-*")
	defer func() { _ = os.RemoveAll(dir) }()

	pth := oskit.WriteStr(t, "hello\n", dir, "out.txt")
	data, _ := os.ReadFile(pth)
	fmt.Print(string(data))
	// Output:
	// hello
}

func ExampleCreateStr() {
	t := &testing.T{}
	dir, _ := os.MkdirTemp("", "oskit-*")
	defer func() { _ = os.RemoveAll(dir) }()

	// CreateStr truncates: a shorter second write shortens the file.
	oskit.CreateStr(t, "abcdef", dir, "f.txt")
	oskit.CreateStr(t, "xy", dir, "f.txt")

	fmt.Println(oskit.ReadFileStr(t, dir, "f.txt"))
	// Output:
	// xy
}

func ExampleMkdirAll() {
	t := &testing.T{}
	dir, _ := os.MkdirTemp("", "oskit-*")
	defer func() { _ = os.RemoveAll(dir) }()

	pth := oskit.MkdirAll(t, dir, "a", "b", "c")
	fi, _ := os.Stat(pth)
	fmt.Println(fi.IsDir())
	// Output:
	// true
}

func ExamplePathExists() {
	t := &testing.T{}
	fmt.Println(oskit.PathExists(t, "testdata/file.txt"))
	fmt.Println(oskit.PathExists(t, "testdata/no_such_file.txt"))
	// Output:
	// true
	// false
}

func ExampleList() {
	t := &testing.T{}
	entries := oskit.List(t, "testdata/list")
	for _, e := range entries {
		fmt.Println(e)
	}
	// Output:
	// d|dir
	// file0.txt
	// file1.txt
}

func ExampleCopyFile() {
	t := &testing.T{}
	dir, _ := os.MkdirTemp("", "oskit-*")
	defer func() { _ = os.RemoveAll(dir) }()

	dst := oskit.CopyFile(t, dir, "testdata/file.txt")
	data, _ := os.ReadFile(dst)
	fmt.Println(string(data))
	// Output:
	// content
}

func ExampleEnvSplit() {
	env := []string{"HOME=/home/user", "EDITOR=vim", "TERM=xterm"}
	m := oskit.EnvSplit(env)
	fmt.Println(m["EDITOR"])
	// Output:
	// vim
}

func ExampleEnvSplitOrdered() {
	env := []string{"A=1", "B=2", "C=3"}
	m, order := oskit.EnvSplitOrdered(env)
	fmt.Println(order)
	fmt.Println(m["B"])
	// Output:
	// [A B C]
	// 2
}

func ExampleEnvJoin() {
	pairs := oskit.EnvJoin(map[string]string{"KEY": "value"})
	fmt.Println(pairs[0])
	// Output:
	// KEY=value
}
