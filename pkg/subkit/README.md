<!-- TOC -->
* [The `subkit` package](#the-subkit-package)
  * [Subprocess test pattern](#subprocess-test-pattern)
    * [New](#new)
    * [NewPkg](#newpkg)
  * [Detection helpers](#detection-helpers)
    * [InSubProcess](#insubprocess)
    * [InMainProcess](#inmainprocess)
  * [Coverage forwarding](#coverage-forwarding)
    * [GetCovProfile](#getcovprofile)
<!-- TOC -->

# The `subkit` package

`subkit` makes it easy to run a single test or an entire package as a
child `go test` process. The typical use case is code that calls
`os.Exit`, `log.Fatal`, or any other function that terminates the
process — behaviour that cannot be tested in-process without aborting
the parent binary.

The package works by setting a per-test sentinel environment variable
when launching the child. Both the parent and child call `New` (or
`NewPkg`) with the same name; `InMainProcess` and `InSubProcess` read
that variable to determine which side of the fork is executing.

## Subprocess test pattern

### New

`New` creates a `SubProcess` targeting a single test by name. In a
real test, pass `t.Name()` to match the enclosing test exactly:

<!-- gmdoceg:ExampleNew -->
```go
sub := subkit.New("Test_MyFunc/success")
if sub.InMainProcess() {
    // sout, eout, err := sub.Run()
    return
}
// --- IN SUBPROCESS ---
// The real test body runs here in the child process.

// Output:
```

`Run` spawns a child with `go test -v -test.run=<name>` and the
sentinel variable set. The child re-enters the function and falls
through to the test body below the guard.

### NewPkg

`NewPkg` creates a `SubProcess` that runs an entire package. The
argument must be an import path accepted by `go test`:

<!-- gmdoceg:ExampleNewPkg -->
```go
sub := subkit.NewPkg("github.com/ctx42/testkit/pkg/myservice")
// sout, eout, err := sub.Run()
_ = sub
// Output:
```

## Detection helpers

### InSubProcess

`InSubProcess` returns true when the current process is the child
launched by `Run`. It checks for the sentinel environment variable
that `Run` sets before spawning the child:

<!-- gmdoceg:ExampleSubProcess_InSubProcess -->
```go
sub := subkit.New("Test_Example")
fmt.Println(sub.InSubProcess())
// Output:
// false
```

### InMainProcess

`InMainProcess` is the inverse of `InSubProcess`. It returns true in
the parent process and is the usual guard condition:

<!-- gmdoceg:ExampleSubProcess_InMainProcess -->
```go
sub := subkit.New("Test_Example")
fmt.Println(sub.InMainProcess())
// Output:
// true
```

## Coverage forwarding

### GetCovProfile

`GetCovProfile` scans an args slice for the `-test.coverprofile` flag
and returns it together with its value as a two-element slice, or nil
when absent. Pass `os.Args` to extract the flag from the current
`go test` invocation and forward it into a subprocess managed with
`exec.Command`:

<!-- gmdoceg:ExampleGetCovProfile -->
```go
args := []string{"-test.v", "-test.coverprofile", "/tmp/cover.out"}
fmt.Println(subkit.GetCovProfile(args))
// Output:
// [-test.coverprofile /tmp/cover.out]
```

When the flag is not present, `GetCovProfile` returns nil:

<!-- gmdoceg:ExampleGetCovProfile_notPresent -->
```go
args := []string{"-test.v", "-test.timeout", "30s"}
fmt.Println(subkit.GetCovProfile(args))
// Output:
// []
```
