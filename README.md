<!-- TOC -->
* [testkit](#testkit)
  * [Installation](#installation)
  * [Packages at a glance](#packages-at-a-glance)
  * [randkit — random test data](#randkit--random-test-data)
  * [timekit — deterministic clocks](#timekit--deterministic-clocks)
  * [iokit — buffers and error injection](#iokit--buffers-and-error-injection)
    * [Thread-safe test buffers](#thread-safe-test-buffers)
    * [Error-injecting readers and writers](#error-injecting-readers-and-writers)
    * [Seek and offset helpers](#seek-and-offset-helpers)
  * [reflectkit — struct field inspection](#reflectkit--struct-field-inspection)
  * [exekit — running external commands](#exekit--running-external-commands)
  * [selfkit — subprocess testing](#selfkit--subprocess-testing)
  * [testkit — global cleanup and hashing](#testkit--global-cleanup-and-hashing)
    * [SHA-1 helpers](#sha-1-helpers)
    * [Global post-test cleanup](#global-post-test-cleanup)
  * [Further reading](#further-reading)
<!-- TOC -->

# testkit

`testkit` is a curated collection of focused Go testing utilities.
Each sub-package solves one problem well: deterministic time, I/O
error injection, subprocess control, random fixture data, and more.
Every helper is designed to integrate naturally with `tester.T` and
the assertion packages in the CTX42 testing module.

The common thread is that testkit helpers catch mistakes at cleanup
time — a `WetBuffer` fails the test if nothing was written; a
`DryBuffer` fails if something was. You write the happy path once
and the invariants enforce themselves.

## Installation

```
go get github.com/ctx42/testkit
```

## Packages at a glance

| Package      | Import path                               | What it does                                                  |
|--------------|-------------------------------------------|---------------------------------------------------------------|
| `randkit`    | `github.com/ctx42/testkit/pkg/randkit`    | Cryptographically random strings, file names, integers        |
| `timekit`    | `github.com/ctx42/testkit/pkg/timekit`    | Deterministic and fixed clocks for time-dependent tests       |
| `iokit`      | `github.com/ctx42/testkit/pkg/iokit`      | Thread-safe test buffers; error-injecting readers and writers |
| `reflectkit` | `github.com/ctx42/testkit/pkg/reflectkit` | Safe struct field and value inspection via reflection         |
| `exekit`     | `github.com/ctx42/testkit/pkg/exekit`     | Run external commands and assert their output and exit code   |
| `selfkit`    | `github.com/ctx42/testkit/pkg/selfkit`    | Subprocess testing by re-invoking the test binary itself      |
| `testkit`    | `github.com/ctx42/testkit/pkg/testkit`    | SHA-1 hashing helpers and global post-test cleanup            |

---

## randkit — random test data

`randkit` generates unpredictable test fixtures using `crypto/rand`
so tests never accidentally pass because they share a hardcoded value.
All functions panic if randomness is unavailable, which is the right
behaviour in tests.

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

## iokit — buffers and error injection

### Thread-safe test buffers

`WetBuffer` and `DryBuffer` register a cleanup hook via `t.Cleanup`
that automatically fails the test if the post-condition is violated.
You do not need to write assertion boilerplate for the common cases.

```go
import "github.com/ctx42/testkit/pkg/iokit"

// WetBuffer: fails the test if nothing was written, OR if the
// contents were never examined via buf.String().
func TestHandler_writes_output(t *testing.T) {
    out := iokit.WetBuffer(t, "stdout")
    Handler(out)
    assert.Equal(t, "expected output\n", out.String())
}

// DryBuffer: fails the test if anything is written.
func TestHandler_silent_on_success(t *testing.T) {
    errOut := iokit.DryBuffer(t, "stderr")
    Handler(io.Discard, errOut)
    // No assertion needed — cleanup does it.
}

// Skip the "must examine" requirement when you only care that
// something was written, not what.
buf := iokit.WetBuffer(t).SkipExamine()
```

### Error-injecting readers and writers

Force I/O errors after a precise byte count to test error-handling
paths without mocking the filesystem.

```go
// Fail after reading 4 bytes.
r := iokit.ErrReader(strings.NewReader("hello world"), 4)
data, err := io.ReadAll(r)
// data = []byte("hell"), err = iokit.ErrRead

// Custom error.
r = iokit.ErrReader(src, 4, iokit.WithReadErr(io.ErrUnexpectedEOF))

// Fail a write after 3 bytes.
w := iokit.ErrWriter(&bytes.Buffer{}, 3)
n, err := w.Write([]byte{0, 1, 2, 3})
// n = 3, err = iokit.ErrWrite

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

## reflectkit — struct field inspection

`reflectkit` is useful when testing code that depends on struct
tags or field metadata — for example, asserting that a generated
type carries the correct JSON or validation tags.

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

## exekit — running external commands

`exekit` wraps `os/exec` with a `tester.T`-aware executor that
handles timeouts, environment setup, and exit-code assertions. It
reports failures through `t.Error` so the full test output is
preserved even when a sub-process misbehaves.

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
exe2 := exekit.New(t, exekit.WithExitCode(1))
exe2.Exe("false")
```

When running under `go test -coverprofile`, use `WithDevOsCov` to
propagate the coverage directory to the sub-process automatically:

```go
exe := exekit.New(t, exekit.WithDevOsCov)
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

## testkit — global cleanup and hashing

### SHA-1 helpers

Convenience wrappers for computing SHA-1 hashes in tests. They panic
on error, which is the right behaviour when unexpected I/O failures
should abort the test immediately.

```go
import "github.com/ctx42/testkit/pkg/testkit"

// Hash a file.
sum := testkit.SHA1File("testdata/golden.bin")
assert.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", sum)

// Hash any io.Reader.
sum = testkit.SHA1Reader(bytes.NewReader(fixture))
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
> caveats before reaching for this API.

---

## Further reading

- [iokit README](pkg/iokit/README.md) — Buffer kind reference and
  error-injection type matrix
- [timekit README](pkg/timekit/README.md) — Clock selection guide
- [selfkit README](pkg/selfkit/README.md) — Complete flag reference
  and subprocess test patterns
- [pkg/exekit godoc](https://pkg.go.dev/github.com/ctx42/testkit/pkg/exekit)
- [pkg/randkit godoc](https://pkg.go.dev/github.com/ctx42/testkit/pkg/randkit)
- [pkg/reflectkit godoc](https://pkg.go.dev/github.com/ctx42/testkit/pkg/reflectkit)
