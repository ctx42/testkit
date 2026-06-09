<!-- TOC -->
* [The `iokit` package](#the-iokit-package)
  * [The `Buffer` Type](#the-buffer-type)
    * [WetBuffer](#wetbuffer)
    * [DryBuffer](#drybuffer)
  * [Error writers and readers](#error-writers-and-readers)
    * [ErrReader](#errreader)
    * [ErrReader — custom error](#errreader--custom-error)
    * [ErrWriter](#errwriter)
  * [Seek and offset helpers](#seek-and-offset-helpers)
    * [Offset](#offset)
    * [Seek](#seek)
    * [ReadAllFromStart](#readallfromstart)
<!-- TOC -->

# The `iokit` package

`iokit` provides I/O helpers for tests: self-checking buffers and a
family of readers and writers that fail on demand. Both let you assert
output and exercise error-handling paths without mocking the
filesystem.

## The `Buffer` Type

`Buffer` is a thread-safe wrapper around `bytes.Buffer`. The
constructors register a `t.Cleanup` hook that fails the test when a
post-condition is violated, so you do not have to write the assertion
boilerplate yourself.

### WetBuffer

A `WetBuffer` fails the test unless it is both written to *and*
examined via `buf.String()` — a guard against output that is produced
but never asserted:

```go
func TestAction(t *testing.T) {
    // --- Given ---
    buf := iokit.WetBuffer(t, "wet-buffer")

    // --- When ---
    Action(buf) // Writes to buf.

    // --- Then ---
    // Fails if Action does not write to buf, or if buf.String is unread.
    assert.Equal(t, "expected output", buf.String())
}
```

Call `SkipExamine` when you only care that something was written, not
what:

```go
buf := iokit.WetBuffer(t, "wet-buffer").SkipExamine()
_, _ = buf.WriteString("data") // No failure for unexamined content.
```

### DryBuffer

A `DryBuffer` fails the test if anything is ever written to it — use
it for the stream that should stay empty:

```go
func TestAction(t *testing.T) {
    // --- Given ---
    buf := iokit.DryBuffer(t, "dry-buffer")

    // --- When ---
    DoSomething(buf) // Must not write to buf.

    // --- Then ---
    // Cleanup fails the test if DoSomething wrote anything.
}
```

## Error writers and readers

These helpers control when, and with what error, the most common I/O
interfaces fail. Each one wraps a real reader or writer and returns the
configured error after a precise byte count:

- `ErrReader` — a failing `io.Reader`.
- `ErrReadCloser` — a failing `io.ReadCloser`.
- `ErrReadSeeker` — a failing `io.ReadSeeker`.
- `ErrReadSeekCloser` — a failing `io.ReadSeekCloser`.
- `ErrWriter` — a failing `io.Writer`.
- `ErrWriteCloser` — a failing `io.WriteCloser`.

Use `WithReadErr`, `WithWriteErr`, `WithCloseErr`, and `WithSeekErr` to
set the error returned by each operation.

### ErrReader

<!-- gmdoceg:ExampleErrReader -->
```go
rdr := strings.NewReader("some text")
rcs := iokit.ErrReader(rdr, 3)

data, err := io.ReadAll(rcs)

fmt.Printf("error: %v\n", err)
fmt.Printf(" data: %s\n", string(data))
// Output:
// error: read error
//  data: som
```

### ErrReader — custom error

<!-- gmdoceg:ExampleErrReader_custom_error -->
```go
mye := errors.New("my error")
rdr := strings.NewReader("some text")
rcs := iokit.ErrReader(rdr, 4, iokit.WithReadErr(mye))

data, err := io.ReadAll(rcs)

fmt.Printf("error: %v\n", err)
fmt.Printf(" data: %s\n", string(data))
// Output:
// error: my error
//  data: some
```

### ErrWriter

<!-- gmdoceg:ExampleErrWriter -->
```go
dst := &bytes.Buffer{}
ce := errors.New("my error")
ew := iokit.ErrWriter(dst, 3, iokit.WithWriteErr(ce))

n, err := ew.Write([]byte{0, 1, 2, 3})

fmt.Printf("    n: %d\n", n)
fmt.Printf("error: %v\n", err)
fmt.Printf("  dst: %v\n", dst.Bytes())
// Output:
//     n: 3
// error: my error
//   dst: [0 1 2]
```

## Seek and offset helpers

These helpers wrap the `io.Seeker` operations that tests reach for
most often. Each one panics on error instead of returning it, which
keeps test setup terse — an unexpected seek failure aborts the test
immediately rather than forcing an error check at every call site.

### Offset

`Offset` reports the current offset of an `io.Seeker`:

```go
off := iokit.Offset(rs)
```

### Seek

`Seek` moves to a position and returns the new offset. The arguments
mirror `io.Seeker.Seek`: an offset and a `whence` — one of
`io.SeekStart`, `io.SeekCurrent`, or `io.SeekEnd`:

```go
// Rewind to the start.
iokit.Seek(rs, 0, io.SeekStart)
```

### ReadAllFromStart

`ReadAllFromStart` reads the entire content of an `io.ReadSeeker` from
offset 0, then restores the original position — handy for inspecting
what a seeker holds mid-test without disturbing the code under test:

```go
data := iokit.ReadAllFromStart(rs)
```
