<!-- TOC -->
* [testkit](#testkit)
  * [Installation](#installation)
  * [Packages at a glance](#packages-at-a-glance)
  * [exekit — running external commands](#exekit--running-external-commands)
  * [iokit — buffers and error injection](#iokit--buffers-and-error-injection)
    * [Thread-safe test buffers](#thread-safe-test-buffers)
    * [Error-injecting readers and writers](#error-injecting-readers-and-writers)
    * [Seek and offset helpers](#seek-and-offset-helpers)
  * [randkit — random test data](#randkit--random-test-data)
  * [reflectkit — struct field inspection](#reflectkit--struct-field-inspection)
  * [selfkit — subprocess testing](#selfkit--subprocess-testing)
  * [subkit — subprocess go test runner](#subkit--subprocess-go-test-runner)
  * [testkit — global cleanup and hashing](#testkit--global-cleanup-and-hashing)
    * [SHA-1 helpers](#sha-1-helpers)
    * [Global post-test cleanup](#global-post-test-cleanup)
  * [timekit — deterministic clocks](#timekit--deterministic-clocks)
  * [Further reading](#further-reading)
<!-- TOC -->

# testkit

`testkit` is a curated collection of focused Go testing utilities.
Each sub-package solves one problem well: deterministic time, I/O
error injection, subprocess control, random fixture data, and more.
Every helper is designed to integrate naturally with
[`tester.T`](https://github.com/ctx42/testing/blob/master/pkg/tester/t.go#L21)
and the assertion packages in the CTX42 testing module.

## Installation

```
go get github.com/ctx42/testkit
```

## Packages at a glance

| Package      | Import path                               | What it does                                                            |
|--------------|-------------------------------------------|-------------------------------------------------------------------------|
| `exekit`     | `github.com/ctx42/testkit/pkg/exekit`     | Run external commands and assert their output and exit code             |
| `iokit`      | `github.com/ctx42/testkit/pkg/iokit`      | Thread-safe test buffers; error-injecting readers and writers           |
| `randkit`    | `github.com/ctx42/testkit/pkg/randkit`    | Random strings, integers, passwords, and file names for test fixtures   |
| `reflectkit` | `github.com/ctx42/testkit/pkg/reflectkit` | Struct field and value inspection via reflection                        |
| `selfkit`    | `github.com/ctx42/testkit/pkg/selfkit`    | Use the test binary as an exec target; assert stdout, stderr, exit code |
| `subkit`     | `github.com/ctx42/testkit/pkg/subkit`     | Run tests in a child go test process; same test detects parent vs child |
| `testkit`    | `github.com/ctx42/testkit/pkg/testkit`    | SHA-1 hashing helpers and global post-test cleanup                      |
| `timekit`    | `github.com/ctx42/testkit/pkg/timekit`    | Deterministic and fixed clocks for time-dependent tests                 |

---

## exekit — running external commands

`exekit` wraps `os/exec` with a `tester.T`-aware executor that handles timeouts,
environment setup, and exit-code assertions. It reports failures through
`t.Error` so the full test output is preserved even when a sub-process
misbehaves.

```go
import "github.com/ctx42/testkit/pkg/exekit"

exe := exekit.New(t,
    exekit.WithTimeout(10*time.Second),
    exekit.WithEnv(append(os.Environ(), "APP_ENV=test")),
)

// Assert stdout only (fails if stderr is non-empty).
out := exe.ExeStdout("go", "env", "GOPATH")

// Assert stderr only (fails if stdout is non-empty).
errMsg := exe.ExeStderr("mybin", "--bad-flag")

// Capture both.
sout, seout := exe.Exe("git", "status", "--short")

// Assert a specific exit code.
exe = exekit.New(t, exekit.WithExitCode(1))
exe.Exe("false")
```

---

## iokit — buffers and error injection

### Thread-safe test buffers

`WetBuffer` and `DryBuffer` register a cleanup hook via `t.Cleanup` that
automatically fails the test if the post-condition is violated. You do not need
to write assertion boilerplate for the common cases.

`WetBuffer` fails the test if nothing was written, or if the contents were
never examined via `buf.String()`:

<!-- gmdoceg:ExampleWetBuffer -->
```go
t := &testing.T{}
buf := iokit.WetBuffer(t, "stdout")
_, _ = buf.WriteString("hello")
fmt.Println(buf.String())
// Output:
// hello
```

`DryBuffer` fails the test if anything is written:

<!-- gmdoceg:ExampleDryBuffer -->
```go
t := &testing.T{}
errOut := iokit.DryBuffer(t, "stderr")
// Pass errOut as io.Writer to code that should produce no output.
// Cleanup calls t.Error if anything is written.
_ = errOut
```

Use `SkipExamine` when you only care that something was written, not what:

<!-- gmdoceg:ExampleBuffer_SkipExamine -->
```go
t := &testing.T{}
buf := iokit.WetBuffer(t, "stdout").SkipExamine()
_, _ = buf.WriteString("data")
// buf.String() need not be called — SkipExamine disables the check.
fmt.Println(buf.Kind())
// Output:
// wet
```

### Error-injecting readers and writers

Force I/O errors after a precise byte count to test error-handling
paths without mocking the filesystem.

```go
// Fail after reading 3 bytes.
rdr := strings.NewReader("some text")
r := iokit.ErrReader(rdr, 3)
data, err := io.ReadAll(r)
// data = []byte("som"), err = iokit.ErrRead

// Custom error.
r = iokit.ErrReader(src, 4, iokit.WithReadErr(io.ErrUnexpectedEOF))

// Fail a write after 3 bytes.
w := iokit.ErrWriter(&bytes.Buffer{}, 3, iokit.WithWriteErr(errors.New("my error")))
n, err := w.Write([]byte{0, 1, 2, 3})
// n = 3, err = "my error"

// ReadCloser, ReadSeeker, ReadSeekCloser, and WriteCloser variants
// follow the same pattern. Use WithCloseErr / WithSeekErr to
// inject errors on those operations independently.
rc := iokit.ErrReadCloser(src, 10, iokit.WithCloseErr(errors.New("close failed")))
```

### Seek and offset helpers

```go
// Current offset of any io.Seeker — panics on error.
off := iokit.Offset(rs)

// Seek and return the new offset — panics on error.
iokit.Seek(rs, 0, io.SeekStart)

// Read the full content from offset 0, restoring the original
// position afterwards.
data := iokit.ReadAllFromStart(rs)
```

---

## randkit — random test data

`randkit` generates random test fixtures using the automatically-seeded
global PRNG from `math/rand/v2`. Tests that rely on hardcoded values
can accidentally pass for the wrong reason; `randkit` eliminates that
class of false confidence.

```go
import "github.com/ctx42/testkit/pkg/randkit"

// Default: 10 random letters [a-zA-Z].
name := randkit.Str()

// Custom character set and length.
token := randkit.Str(randkit.WithChars(randkit.Letters, randkit.Digits), randkit.WithLen(32))

// Random integer in [1, 100].
n := randkit.Int(100)

// Random 16-character password (letters + digits, no specials).
pwd := randkit.Password(16)

// Random file path inside t.TempDir() — prefix "file-", extension ".txt".
path := randkit.FileName(t.TempDir())

// Custom extension.
path = randkit.FileName(t.TempDir(), randkit.WithExt(".json"))
```

Use `WithSeed` when a test must assert exact generated values:

```go
// Output is stable for a given seed — useful for golden-value tests.
name := randkit.Str(randkit.WithSeed(1))      // always "qLKZasgepC"
n    := randkit.Int(100, randkit.WithSeed(1)) // always 32
```

> **Warning:** `WithSeed` is for tests only. Never use it in production
> code — the output is fully predictable from the seed.

---

## reflectkit — struct field inspection

`reflectkit` is useful when testing code that depends on struct tags or field
metadata — for example, asserting that a generated type carries the correct
JSON or validation tags.

```go
import "github.com/ctx42/testkit/pkg/reflectkit"

type Event struct {
    ID        string    `json:"id"         validate:"required"`
    CreatedAt time.Time `json:"created_at"`
}

e := &Event{ID: "evt-1"}

// GetField returns the reflect.StructField (tags, type, index, …).
// Reports via t.Error on any mistake — never panics.
fld := reflectkit.GetField(t, e, "CreatedAt")
assert.Equal(t, "created_at", fld.Tag.Get("json"))

// GetValue returns the reflect.Value of the named field.
// Accepts both pointer and non-pointer structs.
val := reflectkit.GetValue(t, e, "ID")
assert.Equal(t, "evt-1", val.String())
```

---

## selfkit — subprocess testing

`selfkit` solves the problem of testing code that calls `os.Exit`.
Instead of building a separate helper binary, each `go test` binary
can act as its own subprocess target. Wire it into `TestMain` once:

```go
import "github.com/ctx42/testkit/pkg/selfkit"

func TestMain(m *testing.M) {
    runTests, exitCode := selfkit.New().Run(os.Stdout, os.Stderr)
    if runTests {
        os.Exit(m.Run())
    }
    os.Exit(exitCode)
}
```

Then in any test, re-invoke the binary with `exekit` or `exec.Command`
and pass the `selfkit` flags to control what it does:

```go
exe := exekit.New(t, exekit.WithExitCode(42), exekit.WithDevOsCov)
exe.Exe(os.Args[0], "-test.run=^$", "--exitCode", "42")

// Assert that a specific string reaches stdout.
exe2 := exekit.New(t, exekit.WithDevOsCov)
out := exe2.ExeStdout(os.Args[0], "-test.run=^$",
    "--noWrap", "--toStdout", "hello")
assert.Equal(t, "hello", out)
```

Available flags: `--toStdout`, `--toStderr`, `--exitCode`,
`--printEnv`, `--printArgs`, `--noWrap`, `--printToStderr`.
See the [selfkit README](pkg/selfkit/README.md) for the full
reference and flag-combination examples.

---

## subkit — subprocess go test runner

`subkit` makes it easy to run a single test or an entire package as a
child `go test` process. The typical use case is code that calls
`os.Exit`, `log.Fatal`, or otherwise terminates the process —
behavior that cannot be tested in-process without aborting the parent
binary.

Add the subprocess guard at the top of the test, before any code that
could exit:

```go
import "github.com/ctx42/testkit/pkg/subkit"

func Test_MyCommand(t *testing.T) {
    sub := subkit.New(t.Name())
    if sub.InMainProcess() {
        sout, eout, err := sub.Run()
        assert.NoError(t, err)
        assert.Contain(t, "expected output", sout)
        assert.Empty(t, eout)
        return
    }
    // --- IN SUBPROCESS ---
    // Code here runs only inside the child process.
    MyCommand()
}
```

`InMainProcess` returns true in the parent `go test` process. `Run`
spawns a child with `go test -v -test.run=<name>` and a sentinel
environment variable set. The child detects that variable via
`InSubProcess` and falls through to the real test body.

To target a whole package instead of a single test:

```go
sub := subkit.NewPkg("./pkg/myservice")
sout, eout, err := sub.Run()
```

---

## testkit — global cleanup and hashing

### SHA-1 helpers

Convenience wrappers for computing SHA-1 hashes in tests. They panic
on error, which is the right behavior when unexpected I/O failures
should abort the test immediately.

```go
import "github.com/ctx42/testkit/pkg/testkit"

// Hash any io.Reader.
r := strings.NewReader("hello")
sum := testkit.SHA1Reader(r)
// sum == "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"

// Hash a file.
sum = testkit.SHA1File("testdata/golden.bin")
assert.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", sum)
```

### Global post-test cleanup

`AddGlobalCleanup` is for the rare case where cleanup must run after
*all* tests in a package finish — for example, stopping a shared
database container. For per-test cleanup, prefer `t.Cleanup`.

```go
// In TestMain or a top-level init for the test binary.
func TestMain(m *testing.M) {
    container := startDB()
    testkit.AddGlobalCleanup(func() { container.Stop() })

    exitCode := m.Run()
    testkit.RunGlobalCleanups()
    os.Exit(exitCode)
}
```

> **Warning:** `AddGlobalCleanup` uses package-level mutable state
> visible to all goroutines. See the godoc for the full list of
> warnings before reaching for this API.

---

## timekit — deterministic clocks

Production code that calls `time.Now()` directly is hard to test
reliably. The fix is to inject a `func() time.Time` — the same
signature as `time.Now` — and replace it in tests with a clock you
control. `timekit` provides three variants.

```go
import "github.com/ctx42/testkit/pkg/timekit"

base := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

// ClockFixed — always returns the same instant, no matter how many
// times it is called. Perfect for "freeze time" scenarios.
clk := timekit.ClockFixed(base)
fmt.Println(clk()) // 2024-06-01 12:00:00 +0000 UTC
fmt.Println(clk()) // 2024-06-01 12:00:00 +0000 UTC

// TikTak — advances exactly one second on every call. Ideal for
// asserting sequence-dependent time logic without sleeping.
clk = timekit.TikTak(base)
fmt.Println(clk()) // 2024-06-01 12:00:00 +0000 UTC
fmt.Println(clk()) // 2024-06-01 12:00:01 +0000 UTC
fmt.Println(clk()) // 2024-06-01 12:00:02 +0000 UTC

// ClockDeterministic — same as TikTak but with a custom tick.
clk = timekit.ClockDeterministic(base, 15*time.Minute)
fmt.Println(clk()) // 2024-06-01 12:00:00 +0000 UTC
fmt.Println(clk()) // 2024-06-01 12:15:00 +0000 UTC

// ClockStartingAt — advances in real time from a given base instant.
// Useful when the test needs a plausible timestamp in a specific era.
clk = timekit.ClockStartingAt(base)
```

Inject the clock into the type under test:

```go
// Production type.
type Scheduler struct {
    now func() time.Time
}
func NewScheduler(now func() time.Time) *Scheduler { ... }

// In tests.
sched := NewScheduler(timekit.TikTak(base))
```

---

## Further reading

- `exekit` [README](pkg/exekit/README.md)
- `iokit` [README](pkg/iokit/README.md)
- `randkit` [README](pkg/randkit/README.md)
- `reflectkit` [README](pkg/reflectkit/README.md)
- `selfkit` [README](pkg/selfkit/README.md)
- `subkit` [README](pkg/subkit/README.md)
- `testkit` [README](pkg/testkit/README.md)
- `timekit` [README](pkg/timekit/README.md)
