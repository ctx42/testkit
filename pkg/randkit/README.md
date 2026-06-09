<!-- TOC -->
* [The `randkit` package](#the-randkit-package)
  * [String generation](#string-generation)
    * [Str](#str)
  * [File names](#file-names)
    * [FileName](#filename)
  * [Integers](#integers)
    * [Int](#int)
  * [Passwords](#passwords)
    * [Password](#password)
  * [Deterministic output](#deterministic-output)
    * [WithSeed](#withseed)
  * [Character set constants](#character-set-constants)
<!-- TOC -->

# The `randkit` package

`randkit` generates random test fixtures using the automatically-seeded
global PRNG from `math/rand/v2`. Tests that rely on hardcoded values
can accidentally pass for the wrong reason; `randkit` eliminates that
class of false confidence.

## String generation

### Str

`Str` returns a random string. With no options it generates ten
random letters `[a-zA-Z]`:

<!-- gmdoceg:ExampleStr -->
```go
fmt.Println(randkit.Str(randkit.WithSeed(1)))
// Output:
// qLKZasgepC
```

Use `WithLen` to change the length:

<!-- gmdoceg:ExampleStr_withLen -->
```go
fmt.Println(randkit.Str(randkit.WithSeed(1), randkit.WithLen(6)))
// Output:
// qLKZas
```

Use `WithChars` to restrict or expand the character set. The built-in
constants `Letters`, `Uppercase`, `Lowercase`, and `Digits` can be
composed freely:

<!-- gmdoceg:ExampleStr_withChars -->
```go
s := randkit.Str(
    randkit.WithChars(randkit.Digits),
    randkit.WithLen(8),
    randkit.WithSeed(1),
)
fmt.Println(s)
// Output:
// 37790310
```

Use `WithPrefix` and `WithSuffix` (or `WithExt`, an alias for
`WithSuffix`) to add fixed text around the random part:

<!-- gmdoceg:ExampleStr_withPrefixSuffix -->
```go
s := randkit.Str(
    randkit.WithPrefix("test-"),
    randkit.WithSuffix("-end"),
    randkit.WithLen(6),
    randkit.WithSeed(1),
)
fmt.Println(s)
// Output:
// test-qLKZas-end
```

## File names

### FileName

`FileName` returns a random file path inside the given directory. The
default name is seven letters with the prefix `"file-"` and the
extension `".txt"`:

<!-- gmdoceg:ExampleFileName -->
```go
fmt.Println(randkit.FileName("/tmp", randkit.WithSeed(1)))
// Output:
// /tmp/file-qLKZasg.txt
```

Use `WithExt` to change the extension:

<!-- gmdoceg:ExampleFileName_withExt -->
```go
name := randkit.FileName("/tmp", randkit.WithExt(".json"), randkit.WithSeed(1))
fmt.Println(name)
// Output:
// /tmp/file-qLKZasg.json
```

## Integers

### Int

`Int` returns a uniform random integer in the closed range `[1, max]`:

<!-- gmdoceg:ExampleInt -->
```go
fmt.Println(randkit.Int(100, randkit.WithSeed(1)))
// Output:
// 32
```

## Passwords

### Password

`Password` returns an `n`-character string drawn from letters and
digits. No special characters are included:

<!-- gmdoceg:ExamplePassword -->
```go
fmt.Println(randkit.Password(16, randkit.WithSeed(1)))
// Output:
// tSR9avhesITXkYun
```

## Deterministic output

### WithSeed

`WithSeed` replaces the global PRNG with a deterministic ChaCha8
source, making the output reproducible for a given seed. Use it when
a test needs to assert exact generated values:

```go
func TestMyFeature(t *testing.T) {
    name := randkit.Str(randkit.WithSeed(42))
    // name is always "uAfUWlGAxu" for seed 42
    assert.Equal(t, "uAfUWlGAxu", name)
}
```

> **Warning:** `WithSeed` is intended for tests only. Never use it in
> production code — the output is fully predictable from the seed and
> provides no security guarantees whatsoever.

## Character set constants

| Constant    | Value                                                  |
|-------------|--------------------------------------------------------|
| `Uppercase` | `"ABCDEFGHIJKLMNOPQRSTUVWXYZ"`                         |
| `Lowercase` | `"abcdefghijklmnopqrstuvwxyz"`                         |
| `Letters`   | `Lowercase + Uppercase`                                |
| `Digits`    | `"0123456789"`                                         |

Combine them freely with `WithChars`:

```go
// Letters and digits only.
s := randkit.Str(randkit.WithChars(randkit.Letters, randkit.Digits))

// Digits only, length 6.
pin := randkit.Str(randkit.WithChars(randkit.Digits), randkit.WithLen(6))
```
