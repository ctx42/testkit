# prjkit

Package `prjkit` provides a `Project` helper for writing integration
tests that operate on a temporary Go project on disk — with source
files, a Go module, a git history, and optional Docker configuration.

## Lifecycle

A `Project` has two states: **open** for configuration and **closed**
for use. Setup methods — `CreateFile`, `GoModInit`, `GitInitAddAll`,
and similar — require the project to be open. Query and execution
methods — `Compile`, `GitHash`, `ImgRef` — require it to be
closed.

Call `Close` once the project is fully configured. A cleanup
registered by `New` enforces this: if `Close` is not called before
the test ends, the test fails.

## Quick Start

```go
func TestBuild(t *testing.T) {
    prj := prjkit.New(t, t.TempDir())
    prj.CreateFileWith("package main\nfunc main() {}", "main.go")
    prj.GoModInit()
    prj.GitInitAddAll()
    prj.Close()

    bin := prj.Compile()
    _ = bin
}
```

## Creating a Project

Use `New` with an absolute root path and optional configuration
functions:

- `WithProjectCreate` — create the root directory if it does not
  exist.
- `WithProjectEnv(env)` — override the environment used when running
  commands.

## File Operations

```go
prj.CreateFile("config.yaml")
prj.CreateFileWith("key=value\n", "configs", "app.conf")
prj.FilesFrom("testdata/fixtures")
prj.Rename("old.go", "new.go")
```

## Go Module

```go
prj.GoModInit()  // writes go.mod using the default module name
prj.GoModTidy()  // runs go mod tidy
```

The default module name is `example.com/comp/<dir>`, where `<dir>` is
the base name of the project root directory.

## Git

```go
prj.GitInitAddAll("v1.0.0")           // init, stage all, commit, tag
prj.GitSetRemote()                    // add default origin remote
prj.GitCommit("v1.1.0", "Release v1.1.0")

prj.Close()

hash := prj.GitHash()
log  := prj.GitCommitLog()
```

## Docker Configuration

```go
prj.CfgRegRepoDef()          // add default Docker registry to config
prj.CfgBldTargets("app,worker")
prj.WithDockerfile()          // copy example Dockerfile; assign
                              // random image name and tag

prj.Close()

ref    := prj.ImgRef()
latest := prj.ImgRefLatest()
```

## Compile

`Compile` builds the project and returns the absolute path to the
binary. The binary is placed in a temporary directory that is cleaned
up when the test ends.

```go
prj.Close()
bin := prj.Compile()
```

## Parsing git log lines

`NewGitCommit` parses a single `git log --pretty=%h (%D) %at %s`
line into a `GitCommit`:

<!-- gmdoceg:ExampleNewGitCommit -->
```go
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
```

## GitCommits

`GitCommits` is an ordered slice of `GitCommit` (most-recent first).

### Latest

<!-- gmdoceg:ExampleGitCommits_Latest -->
```go
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
```

### First

<!-- gmdoceg:ExampleGitCommits_First -->
```go
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
```

### Find

<!-- gmdoceg:ExampleGitCommits_Find -->
```go
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
```

### N

<!-- gmdoceg:ExampleGitCommits_N -->
```go
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
```
