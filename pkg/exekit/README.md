<!-- TOC -->
* [The `exekit` package](#the-exekit-package)
  * [Running commands](#running-commands)
    * [New](#new)
    * [Exe](#exe)
    * [ExeStdout](#exestdout)
    * [ExeStderr](#exestderr)
  * [Options](#options)
  * [Coverage helpers](#coverage-helpers)
    * [IsWithCoverage](#iswithcoverage)
    * [MaybeAddGoCovDir](#maybeaddgocovdir)
    * [WithDetCov and WithDevOsCov](#withdetcov-and-withdevoscov)
<!-- TOC -->

# The `exekit` package

`exekit` provides test helpers for running external commands and asserting
their exit code, standard output, and standard error. The central type is `Exe`,
which wraps `os/exec` with a configurable timeout, environment, exit-code
expectation, and working directory.

## Running commands

### New

`New` constructs an `Exe` configured with the given options. By default, it
uses `os.Environ()` and a five-second timeout:

```go
exe := exekit.New(t,
    exekit.WithTimeout(10*time.Second),
    exekit.WithEnv(append(os.Environ(), "APP_ENV=test")),
)
```

### Exe

`Exe` runs the command and returns both stdout and stderr. It marks the test as
failed if the command exits with a code other than the expected one:

```go
sout, eout := exe.Exe("git", "status", "--short")
assert.Equal(t, "", eout)
assert.Contain(t, "main.go", sout)
```

### ExeStdout

`ExeStdout` runs the command and returns stdout. It additionally fails the test
if anything was written to stderr:

```go
out := exe.ExeStdout("go", "env", "GOPATH")
assert.Equal(t, gopath, strings.TrimSpace(out))
```

### ExeStderr

`ExeStderr` runs the command and returns stderr. It additionally fails the test
if anything was written to stdout:

```go
msg := exe.ExeStderr("mybin", "--bad-flag")
assert.Contain(t, "unknown flag", msg)
```

## Options

| Option          | Description                                            |
|-----------------|--------------------------------------------------------|
| `WithTimeout`   | Override the execution timeout (default: 5s).          |
| `WithEnv`       | Replace the process environment (slice form).          |
| `WithEnvMap`    | Replace the process environment (map form).            |
| `WithWd`        | Set the working directory for the command.             |
| `WithExitCode`  | Expect a non-zero exit code from the command.          |
| `WithTrim`      | Trim leading/trailing whitespace from both outputs.    |
| `WithDevOsCov`  | Propagate coverage instrumentation to the subprocess.  |

To assert a non-zero exit code:

```go
exe := exekit.New(t, exekit.WithExitCode(1))
exe.Exe("false")
```

## Coverage helpers

### IsWithCoverage

`IsWithCoverage` inspects an args slice to determine whether tests were run
with coverage enabled. It recognizes both `-coverprofile` and
`-test.coverprofile` in `=value` and space-separated forms:

<!-- gmdoceg:ExampleIsWithCoverage -->
```go
args := []string{"-test.v", "-test.coverprofile=/tmp/cover.out"}
path, ok := IsWithCoverage(args)
fmt.Println(path)
fmt.Println(ok)
// Output:
// /tmp/cover.out
// true
```

When the flag is absent, `IsWithCoverage` returns `("", false)`:

<!-- gmdoceg:ExampleIsWithCoverage_absent -->
```go
args := []string{"-test.v", "-test.timeout=30s"}
_, ok := IsWithCoverage(args)
fmt.Println(ok)
// Output:
// false
```

### MaybeAddGoCovDir

`MaybeAddGoCovDir` adds `GOCOVERDIR` to an environment slice when coverage is
enabled in the given args. It is a no-op when coverage is not detected or when
`GOCOVERDIR` is already present:

<!-- gmdoceg:ExampleMaybeAddGoCovDir -->
```go
env := []string{"HOME=/root"}
args := []string{"-test.coverprofile=/tmp/cover.out"}
env = MaybeAddGoCovDir(env, args, func() string { return "/tmp/covdir" })
fmt.Println(env)
// Output:
// [HOME=/root GOCOVERDIR=/tmp/covdir]
```

### WithDetCov and WithDevOsCov

`WithDetCov` is an `Exe` option that adds `GOCOVERDIR` to the executor's
environment when the provided args contain a coverage flag. `WithDevOsCov` is a
convenience wrapper that reads `os.Args` automatically — apply it last so it
sees the final environment:

```go
exe := exekit.New(t, exekit.WithDevOsCov)
sout, _ := exe.Exe(os.Args[0], "-test.run=^$")
```
