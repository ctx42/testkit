<!-- TOC -->
* [The `iokit` package](#the-iokit-package)
  * [The `Buffer` Type](#the-buffer-type)
    * [WetBuffer](#wetbuffer)
    * [DryBuffer](#drybuffer)
  * [Error writers and readers](#error-writers-and-readers)
    * [ErrReader](#errreader)
    * [ErrReader — custom error](#errreader--custom-error)
    * [ErrWriter](#errwriter)
<!-- TOC -->

# The `iokit` package

The `iokit` package provides I/O and buffer related helpers.

## The `Buffer` Type

The `Buffer` type, defined in `buffer.go`, is a thread-safe wrapper around
`bytes.Buffer`. It supports three kinds of behavior for test cleanup:

### WetBuffer

A `WetBuffer` uses `Buffer` and ensures it's written to and its contents are examined during the test.

```go
func TestAction(t *testing.T) {
    // --- Given ---
    buf := iokit.WetBuffer(t, "wet-buffer")

    // --- When ---
    Action(buf) // Writes to buf.

    // --- Then ---
    // Fails if Action doesn't write to buf or if buf.String() is not called.
    assert.Equal(t, "expected output", buf.String())
}
```

To skip the examination requirement:

```go
buf := iokit.WetBuffer(t, "wet-buffer").SkipExamine()
buf.WriteString("data") // No failure for unexamined content
```

### DryBuffer

A DryBuffer ensures the buffer remains empty.

```go
func TestAction(t *testing.T) {
    // --- Given ---
    buf := iokit.DryBuffer(t, "dry-buffer")

    // --- When ---
    DoSomething(buf) // Must not write to buf.

    // --- Then ---
    // Fails if DoSomething writes to buf.
}
```

## Error writers and readers

Package provides helpers for controlling when and how the most important
I/O interfaces return an error.

- `ErrReader` — control when and what error an `io.Reader` returns.
- `ErrReadCloser` — control when and what error an `io.ReadCloser` returns.
- `ErrReadSeeker` — control when and what error an `io.ReadSeeker` returns.
- `ErrReadSeekCloser` — control when and what error an `io.ReadSeekCloser` returns.
- `ErrWriter` — control when and what error an `io.Writer` returns.
- `ErrWriteCloser` — control when and what error an `io.WriteCloser` returns.

### ErrReader

<!-- gmdoceg:ExampleErrReader -->
```go
rdr := strings.NewReader("some text")
rcs := ErrReader(rdr, 3)

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
rcs := ErrReader(rdr, 4, WithReadErr(mye))

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
ew := ErrWriter(dst, 3, WithWriteErr(ce))

n, err := ew.Write([]byte{0, 1, 2, 3})

fmt.Printf("    n: %d\n", n)
fmt.Printf("error: %v\n", err)
fmt.Printf("  dst: %v\n", dst.Bytes())
// Output:
//     n: 3
// error: my error
//   dst: [0 1 2]
```
