<!-- TOC -->
* [The `selfkit` package](#the-selfkit-package)
  * [Setup](#setup)
  * [Writing a subprocess test](#writing-a-subprocess-test)
  * [Flags](#flags)
    * [--toStdout and --toStderr](#--tostdout-and---tostderr)
    * [--exitCode](#--exitcode)
    * [--printEnv](#--printenv)
    * [--printArgs](#--printargs)
    * [--noWrap](#--nowrap)
    * [--printToStderr](#--printtostderr)
<!-- TOC -->

# The `selfkit` package

The `selfkit` package lets you test code that calls `os.Exit` — or
any code you need to run in a subprocess — by re-invoking the test
binary itself. Rather than building a separate helper binary, each
`go test` binary can act as its own controlled subprocess target.

`selfkit` works through a small set of CLI flags. When the test binary
is re-invoked with one of those flags, `Self.Run` handles the request
(prints output, sets the exit code) and tells `TestMain` to exit
without running tests.

## Setup

Wire `selfkit` into `TestMain` once per package that needs it:

```go
func TestMain(m *testing.M) {
    runTests, exitCode := selfkit.NewT().Run(os.Stdout, os.Stderr)
    if runTests {
        os.Exit(m.Run())
    }
    os.Exit(exitCode)
}
```

When no selfkit flags are present (normal test run), `Run` returns
`(true, 0)` and the caller proceeds with `m.Run()`:

<!-- gmdoceg:ExampleNew -->
```go
// No selfkit flags: Run signals the caller to proceed with tests.
se := selfkit.NewT(selfkit.WithArgs([]string{"prog"}))

runTests, exitCode := se.Run(io.Discard, io.Discard)

fmt.Println(runTests)
fmt.Println(exitCode)
// Output:
// true
// 0
```

## Writing a subprocess test

Use `exec.Command(os.Args[0], ...)` to re-invoke the test binary.
Pass `-test.run=^$` to suppress any actual test execution in the
subprocess, then add the selfkit flag that controls its behaviour:

```go
func Test_MyCode_exits(t *testing.T) {
    var stdout, stderr bytes.Buffer

    cmd := exec.Command(os.Args[0],
        "-test.run=^$",
        "--exitCode", "42",
    )
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()

    var exitErr *exec.ExitError
    if !errors.As(err, &exitErr) {
        t.Fatalf("expected exit error, got %v", err)
    }
    assert.Equal(t, 42, exitErr.ExitCode())
}
```

## Flags

### --toStdout and --toStderr

Print a value to stdout or stderr and exit without running tests.
The output is wrapped in a sentinel so the caller can distinguish
it from other binary noise:

```
--toStdout VALUE  →  |sout: VALUE|
--toStderr VALUE  →  |eout: VALUE|
```

Both flags may be combined in one invocation:

<!-- gmdoceg:ExampleSelf_Run_toStdout -->
```go
var sout, eout strings.Builder
args := []string{"prog", "--toStdout", "hello", "--toStderr", "world"}

se := selfkit.NewT(selfkit.WithArgs(args))
testsRun, _ := se.Run(&sout, &eout)

fmt.Println(sout.String())
fmt.Println(eout.String())
fmt.Println(testsRun)
// Output:
// |sout: hello|
// |eout: world|
// false
```

### --exitCode

Exit with the given code without running tests. Any integer value
including 0 is accepted; 0 explicitly means "exit success without
running tests".

<!-- gmdoceg:ExampleSelf_Run_exitCode -->
```go
args := []string{"prog", "--exitCode", "42"}

se := selfkit.NewT(selfkit.WithArgs(args))
runTests, exitCode := se.Run(io.Discard, io.Discard)

fmt.Println(runTests)
fmt.Println(exitCode)
// Output:
// false
// 42
```

### --printEnv

Read an environment variable by name and print its value to stdout
(wrapped as `|env: VALUE|`). Useful for asserting that a subprocess
inherits or sets a particular environment variable.

<!-- gmdoceg:ExampleSelf_Run_printEnv -->
```go
var sout strings.Builder
_ = os.Setenv("MY_VAR", "secret")
defer func() { _ = os.Unsetenv("MY_VAR") }()
args := []string{"prog", "--printEnv", "MY_VAR"}

se := selfkit.NewT(selfkit.WithArgs(args))
testsRun, _ := se.Run(&sout, io.Discard)

fmt.Println(testsRun)
fmt.Println(sout.String())
// Output:
// false
// |env: secret|
```

### --printArgs

Print a label followed by any remaining (non-flag) arguments to
stdout, wrapped as `|args: LABEL REST1,REST2|`. Put `--printArgs`
last so trailing positional arguments are captured correctly.

<!-- gmdoceg:ExampleSelf_Run_printArgs -->
```go
var sout strings.Builder
args := []string{"prog", "--printArgs", "label", "arg1", "arg2"}

se := selfkit.NewT(selfkit.WithArgs(args))
se.Run(&sout, io.Discard)

fmt.Println(sout.String())
// Output:
// |args: label arg1,arg2|
```

### --noWrap

Remove the `|prefix: ...|` sentinel wrappers from all output. Use
this when the caller needs the raw value without delimiters.

<!-- gmdoceg:ExampleSelf_Run_noWrap -->
```go
var sout strings.Builder
args := []string{"prog", "--noWrap", "--toStdout", "hello"}

se := selfkit.NewT(selfkit.WithArgs(args))
se.Run(&sout, io.Discard)

fmt.Println(sout.String())
// Output:
// hello
```

### --printToStderr

Redirect the output of `--printEnv` and `--printArgs` to stderr
instead of stdout.

<!-- gmdoceg:ExampleSelf_Run_printToStderr -->
```go
_ = os.Setenv("MY_VAR", "value")
defer func() { _ = os.Unsetenv("MY_VAR") }()
var eout strings.Builder
args := []string{"prog", "--printToStderr", "--printEnv", "MY_VAR"}

se := selfkit.NewT(selfkit.WithArgs(args))
se.Run(io.Discard, &eout)

fmt.Println(eout.String())
// Output:
// |env: value|
```
