<!-- TOC -->
* [The `testkit` package](#the-testkit-package)
  * [SHA-1 hashing](#sha-1-hashing)
    * [SHA1Reader](#sha1reader)
    * [SHA1File](#sha1file)
  * [Global cleanup](#global-cleanup)
    * [AddGlobalCleanup and RunGlobalCleanups](#addglobalcleanup-and-runglobalcleanups)
<!-- TOC -->

# The `testkit` package

`testkit` provides miscellaneous test helpers that do not belong to a
more focused sub-package: SHA-1 convenience wrappers for asserting
file and stream content, and a global post-test cleanup mechanism for
shared resources that outlive individual tests.

## SHA-1 hashing

Both hashing helpers panic on I/O errors, which is the correct
behaviour when unexpected failures should abort the test immediately.

### SHA1Reader

`SHA1Reader` computes the SHA-1 hash of everything in an `io.Reader`
and returns it as a lowercase hex string:

<!-- gmdoceg:ExampleSHA1Reader -->
```go
r := strings.NewReader("hello")
fmt.Println(testkit.SHA1Reader(r))
// Output:
// aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
```

### SHA1File

`SHA1File` opens the file at the given path and returns its SHA-1
hash. Use it to assert golden-file content without reading the file
manually:

```go
sum := testkit.SHA1File("testdata/golden.bin")
assert.Equal(t, "da39a3ee5e6b4b0d3255bfef95601890afd80709", sum)
```

## Global cleanup

### AddGlobalCleanup and RunGlobalCleanups

`AddGlobalCleanup` registers a function to run after all tests in the
package have finished. Wire both calls into `TestMain`:

<!-- gmdoceg:ExampleAddGlobalCleanup -->
```go
testkit.AddGlobalCleanup(func() {
    // cleanup logic, e.g. stopping a shared database container
})
testkit.RunGlobalCleanups()
// Output:
```

A typical `TestMain` looks like this:

```go
func TestMain(m *testing.M) {
    container := startDB()
    testkit.AddGlobalCleanup(func() { container.Stop() })

    exitCode := m.Run()
    testkit.RunGlobalCleanups()
    os.Exit(exitCode)
}
```

`RunGlobalCleanups` logs each registered cleanup (file and line) to
stderr before invoking it. The log format is:

```
*** TESTKIT running global cleanup function registered in foo_test.go:42
```

> **Warning:** `AddGlobalCleanup` uses package-level mutable state
> visible to all goroutines. Prefer `t.Cleanup` or `t.TempDir` for
> per-test cleanup whenever possible.
