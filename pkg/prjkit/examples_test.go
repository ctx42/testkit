// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package prjkit_test

import (
	"fmt"

	"github.com/ctx42/testkit/pkg/prjkit"
)

func ExampleNewGitCommit() {
	gc, err := prjkit.NewGitCommit("16c806f () 946782245 commit 1")
	if err != nil {
		panic(err)
	}
	fmt.Println(gc.Hash)
	fmt.Println(gc.Summary)
	fmt.Println(gc.Date)
	// Output:
	// 16c806f
	// commit 1
	// 2000-01-02 03:04:05 +0000 UTC
}

func ExampleGitCommits_Latest() {
	commits := prjkit.GitCommits{
		{Hash: "abc1234", Summary: "second commit"},
		{Hash: "def5678", Summary: "first commit"},
	}
	cm := commits.Latest()
	fmt.Println(cm.Hash)
	fmt.Println(cm.Summary)
	// Output:
	// abc1234
	// second commit
}

func ExampleGitCommits_First() {
	commits := prjkit.GitCommits{
		{Hash: "abc1234", Summary: "second commit"},
		{Hash: "def5678", Summary: "first commit"},
	}
	cm := commits.First()
	fmt.Println(cm.Hash)
	fmt.Println(cm.Summary)
	// Output:
	// def5678
	// first commit
}

func ExampleGitCommits_Find() {
	commits := prjkit.GitCommits{
		{Hash: "abc1234", Summary: "second commit"},
		{Hash: "def5678", Summary: "first commit"},
	}
	cm := commits.Find("def5678")
	fmt.Println(cm.Hash)
	fmt.Println(cm.Summary)
	// Output:
	// def5678
	// first commit
}

func ExampleGitCommits_N() {
	commits := prjkit.GitCommits{
		{Hash: "abc1234", Summary: "second commit"},
		{Hash: "def5678", Summary: "first commit"},
	}
	cm := commits.N(1)
	fmt.Println(cm.Hash)
	fmt.Println(cm.Summary)
	// Output:
	// def5678
	// first commit
}
