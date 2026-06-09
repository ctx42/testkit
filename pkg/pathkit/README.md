<!-- TOC -->
* [The `pathkit` package](#the-pathkit-package)
  * [Path helpers](#path-helpers)
    * [AbsPath](#abspath)
    * [EvalSymlinks](#evalsymlinks)
<!-- TOC -->

# The `pathkit` package

`pathkit` provides test helpers for common [path/filepath] operations.
Every exported function integrates with `tester.T`: on error it marks
the test as failed, writes a diagnostic to the test log, and returns an
empty string so the calling test can continue executing.

## Path helpers

### AbsPath

`AbsPath` wraps [filepath.Abs]. Path segments are joined with
[filepath.Join], so you can pass the directory and file name
separately. It returns the absolute path — or an empty string and a
failed test if the path cannot be resolved:

<!-- gmdoceg:ExampleAbsPath -->
```go
t := &testing.T{}

// AbsPath joins the segments, then resolves them to an absolute path.
abs := pathkit.AbsPath(t, ".", "testdata")

fmt.Println(filepath.IsAbs(abs))
fmt.Println(filepath.Base(abs))
// Output:
// true
// testdata
```

### EvalSymlinks

`EvalSymlinks` wraps [filepath.EvalSymlinks], resolving any symbolic
links in the joined path to their real target. It returns an empty
string and fails the test if the path cannot be resolved:

<!-- gmdoceg:ExampleEvalSymlinks -->
```go
t := &testing.T{}

// testdata/dir_sym_link is a symlink to testdata/dir; EvalSymlinks
// resolves it to the real path.
resolved := pathkit.EvalSymlinks(t, "testdata", "dir_sym_link")

fmt.Println(resolved)
// Output:
// testdata/dir
```
