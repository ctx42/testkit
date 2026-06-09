<!-- TOC -->
* [The `oskit` package](#the-oskit-package)
  * [File and directory helpers](#file-and-directory-helpers)
    * [ReadFile / ReadFileStr](#readfile--readfilestr)
    * [WriteStr / Write](#writestr--write)
    * [CreateStr / Create](#createstr--create)
    * [MkdirAll](#mkdirall)
    * [MkdirTemp](#mkdirtemp)
    * [Open](#open)
    * [PathExists](#pathexists)
    * [Stat / FileSize / ModTime / ModTimeSet](#stat--filesize--modtime--modtimeset)
    * [List / ListAbs / Readdirnames](#list--listabs--readdirnames)
    * [CopyFile / CopyDir](#copyfile--copydir)
  * [Working directory helpers](#working-directory-helpers)
    * [Getwd](#getwd)
    * [Chdir](#chdir)
  * [Environment helpers](#environment-helpers)
    * [Setenv / Unsetenv](#setenv--unsetenv)
    * [EnvSplit / EnvSplitOrdered / EnvJoin](#envsplit--envsplitordered--envjoin)
<!-- TOC -->

# The `oskit` package

`oskit` provides test helpers for common [os] operations. Every
exported function integrates with `tester.T`: on error it marks the
test as failed, writes a diagnostic to the test log, and returns a
safe zero value so the calling test can continue executing.

## File and directory helpers

### ReadFile / ReadFileStr

Wrappers around [os.ReadFile]. Path segments are joined with
[filepath.Join], so you can pass the directory and the file name
separately. `ReadFileStr` returns a `string` instead of `[]byte`.

<!-- gmdoceg:ExampleReadFile -->
```go
t := &testing.T{}
data := oskit.ReadFile(t, "testdata/file.txt")
fmt.Println(string(data))
// Output:
// content
```

<!-- gmdoceg:ExampleReadFileStr -->
```go
t := &testing.T{}
s := oskit.ReadFileStr(t, "testdata/file.txt")
fmt.Println(s)
// Output:
// content
```

### WriteStr / Write

`WriteStr` (and its `[]byte` twin `Write`) appends content to a file,
creating it if it does not exist. Multiple calls to `WriteStr` on the
same path accumulate content.

<!-- gmdoceg:ExampleWriteStr -->
```go
t := &testing.T{}
dir, _ := os.MkdirTemp("", "oskit-*")
defer func() { _ = os.RemoveAll(dir) }()

pth := oskit.WriteStr(t, "hello\n", dir, "out.txt")
data, _ := os.ReadFile(pth)
fmt.Print(string(data))
// Output:
// hello
```

### CreateStr / Create

`CreateStr` (and its `[]byte` twin `Create`) writes content to a file
from offset 0, truncating any existing content. Unlike `WriteStr`, a
second call with a shorter string will shorten the file.

<!-- gmdoceg:ExampleCreateStr -->
```go
t := &testing.T{}
dir, _ := os.MkdirTemp("", "oskit-*")
defer func() { _ = os.RemoveAll(dir) }()

// CreateStr truncates: a shorter second write shortens the file.
oskit.CreateStr(t, "abcdef", dir, "f.txt")
oskit.CreateStr(t, "xy", dir, "f.txt")

fmt.Println(oskit.ReadFileStr(t, dir, "f.txt"))
// Output:
// xy
```

### MkdirAll

Creates a directory (and all missing parents) with `0755` permissions.
If the path already exists, it is a no-op. Always returns the
constructed path.

<!-- gmdoceg:ExampleMkdirAll -->
```go
t := &testing.T{}
dir, _ := os.MkdirTemp("", "oskit-*")
defer func() { _ = os.RemoveAll(dir) }()

pth := oskit.MkdirAll(t, dir, "a", "b", "c")
fi, _ := os.Stat(pth)
fmt.Println(fi.IsDir())
// Output:
// true
```

### MkdirTemp

Wraps [os.MkdirTemp] and registers a cleanup that removes the
directory when the test ends. Unlike `t.TempDir`, you choose the
parent directory explicitly.

```go
// Temp dir inside another temp dir.
inner := oskit.MkdirTemp(t, t.TempDir(), "prefix-*")
```

### Open

Wraps [os.Open] and registers a cleanup that closes the file
descriptor when the test ends.

```go
fil := oskit.Open(t, "testdata", "golden.bin")
data, _ := io.ReadAll(fil)
```

### PathExists

Reports whether a path exists on the filesystem, regardless of its
type (file, directory, symlink, etc.).

<!-- gmdoceg:ExamplePathExists -->
```go
t := &testing.T{}
fmt.Println(oskit.PathExists(t, "testdata/file.txt"))
fmt.Println(oskit.PathExists(t, "testdata/no_such_file.txt"))
// Output:
// true
// false
```

### Stat / FileSize / ModTime / ModTimeSet

Thin wrappers around [os.Stat] and [os.Chtimes] for common assertions.

```go
// os.FileInfo for the path.
fi := oskit.Stat(t, "testdata/file.txt")
assert.Equal(t, int64(7), fi.Size())

// File size in bytes.
n := oskit.FileSize(t, "testdata/file.txt")

// Modification time in UTC.
mt := oskit.ModTime(t, "testdata/file.txt")

// Pin modification time for reproducible tests.
oskit.ModTimeSet(t, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), pth)
```

### List / ListAbs / Readdirnames

`List` returns directory entries as `name` or `d|name` (the `d|`
prefix distinguishes directories from files). Entries are sorted.
`ListAbs` returns the same list with absolute paths.
`Readdirnames` returns only bare names, sorted.

<!-- gmdoceg:ExampleList -->
```go
t := &testing.T{}
entries := oskit.List(t, "testdata/list")
for _, e := range entries {
    fmt.Println(e)
}
// Output:
// d|dir
// file0.txt
// file1.txt
```

### CopyFile / CopyDir

`CopyFile` copies a single regular file into a destination directory,
preserving the base name. `CopyDir` copies a directory tree
recursively; symlinks are followed and their targets are copied as
regular files.

<!-- gmdoceg:ExampleCopyFile -->
```go
t := &testing.T{}
dir, _ := os.MkdirTemp("", "oskit-*")
defer func() { _ = os.RemoveAll(dir) }()

dst := oskit.CopyFile(t, dir, "testdata/file.txt")
data, _ := os.ReadFile(dst)
fmt.Println(string(data))
// Output:
// content
```

```go
// Copy a whole directory tree.
oskit.CopyDir(t, t.TempDir(), "testdata/dir")
```

---

## Working directory helpers

### Getwd

Wraps [os.Getwd] and fails the test on error instead of returning an
error value.

```go
wd := oskit.Getwd(t) // panics never; fails t on error
```

### Chdir

Changes the working directory and restores it when the test ends. A
cleanup is registered via `t.Cleanup`, so parallel tests are safe as
long as they do not call `Chdir` concurrently (the function will panic
in that case — by design, since parallel `Chdir` is always a race).

```go
orig := oskit.Getwd(t)
oskit.Chdir(t, "testdata", "dir")
// working directory is now testdata/dir
// restored to orig when t ends
```

---

## Environment helpers

### Setenv / Unsetenv

`Setenv` sets an environment variable. `Unsetenv` unsets one and
restores its previous value when the test ends.

```go
oskit.Setenv(t, "MY_FLAG", "1")
oskit.Unsetenv(t, "MY_FLAG") // restores original value on cleanup
```

### EnvSplit / EnvSplitOrdered / EnvJoin

Utilities for working with the `[]string` slices that [os.Environ]
and `exec.Cmd.Env` use.

`EnvSplit` converts to a `map[string]string`. `EnvSplitOrdered`
returns the map and the original key order. `EnvJoin` converts back to
`KEY=value` strings.

<!-- gmdoceg:ExampleEnvSplit -->
```go
env := []string{"HOME=/home/user", "EDITOR=vim", "TERM=xterm"}
m := oskit.EnvSplit(env)
fmt.Println(m["EDITOR"])
// Output:
// vim
```

<!-- gmdoceg:ExampleEnvSplitOrdered -->
```go
env := []string{"A=1", "B=2", "C=3"}
m, order := oskit.EnvSplitOrdered(env)
fmt.Println(order)
fmt.Println(m["B"])
// Output:
// [A B C]
// 2
```

<!-- gmdoceg:ExampleEnvJoin -->
```go
pairs := oskit.EnvJoin(map[string]string{"KEY": "value"})
fmt.Println(pairs[0])
// Output:
// KEY=value
```
