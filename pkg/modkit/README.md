<!-- TOC -->
* [The `modkit` package](#the-modkit-package)
  * [Module location](#module-location)
    * [Root](#root)
    * [Path](#path)
    * [Tmp](#tmp)
  * [Reading go.mod](#reading-gomod)
    * [Ver](#ver)
    * [ModVer](#modver)
    * [GoVer](#gover)
<!-- TOC -->

# The `modkit` package

`modkit` provides test helpers for working with the Go module under
test: finding the module root and reading versions out of `go.mod`
files.

The helpers fail in three different ways, by design. `Root`, `Path`,
and `Ver` resolve paths against the module root and panic on failure,
since a test cannot proceed without it. `ModVer` and `GoVer` read an
explicit `go.mod` path and return an error instead. `Tmp` integrates
with `tester.T`, reporting failures through the test handle.

## Module location

### Root

`Root` walks up from the current working directory until it finds the
directory containing `go.mod`, and returns its absolute path. It
panics if no `go.mod` is found:

<!-- gmdoceg:ExampleRoot -->
```go
// Root walks up from the working directory to the module root (the
// directory that holds go.mod).
fmt.Println(filepath.Base(modkit.Root()))
// Output:
// testkit
```

### Path

`Path` joins its segments onto the module root and returns the
absolute path. The target need not exist — it is pure path
construction rooted at the module. With no segments it returns the
module root itself:

<!-- gmdoceg:ExamplePath -->
```go
// Path joins its segments onto the module root; the target need not
// exist.
fmt.Println(filepath.Base(modkit.Path("build")))
// Output:
// build
```

### Tmp

`Tmp` creates a directory under `<module-root>/tmp` and registers a
`t.Cleanup` that removes it when the test ends. The first segment is
required; empty segments fail the test. It returns the created path,
or an empty string on error:

```go
// Create <module-root>/tmp/cache/blobs and remove it on cleanup.
dir := modkit.Tmp(t, "cache", "blobs")
```

## Reading go.mod

### Ver

`Ver` returns the version pinned for a module in the current module's
own `go.mod` (located via `Root`). It panics if the `go.mod` cannot be
found or the module is absent or ambiguous:

```go
// Version of a dependency pinned in this module's go.mod.
ver := modkit.Ver("github.com/ctx42/testing")
```

### ModVer

`ModVer` reads an explicit `go.mod` path and returns the version pinned
for the named module. It returns an error if the file cannot be read,
the module is not required, or more than one line matches:

<!-- gmdoceg:ExampleModVer -->
```go
// ModVer reports the version a go.mod pins for a given module.
ver, _ := modkit.ModVer(
    "testdata/mod/go.mod.example",
    "github.com/dave/jennifer",
)
fmt.Println(ver)
// Output:
// v1.5.0
```

### GoVer

`GoVer` returns the Go version declared by the `go` directive in the
`go.mod` at the joined path. It returns an error if the file is
unreadable or does not contain exactly one `go` directive:

<!-- gmdoceg:ExampleGoVer -->
```go
// GoVer reports the Go version declared in a go.mod file.
ver, _ := modkit.GoVer("testdata/mod/go.mod.example")
fmt.Println(ver)
// Output:
// 1.17
```
