<!-- TOC -->
* [The `reflectkit` package](#the-reflectkit-package)
  * [Struct field inspection](#struct-field-inspection)
    * [GetField](#getfield)
    * [GetValue](#getvalue)
<!-- TOC -->

# The `reflectkit` package

`reflectkit` provides lightweight reflection utilities for testing. It
is useful when asserting that a type carries the correct struct tags or
field metadata — for example, verifying that a generated type exposes
the right JSON or validation tags without importing the full struct
definition.

Both functions report errors through `t.Error` and never panic,
keeping test output readable when the struct layout changes.

## Struct field inspection

### GetField

`GetField` returns the `reflect.StructField` for the named field in a
pointer-to-struct. Use it to assert tags, kinds, or other type
metadata:

<!-- gmdoceg:ExampleGetField -->
```go
type Event struct {
    ID    string `json:"id" validate:"required"`
    Score int    `json:"score"`
}
e := &Event{}
fld := GetField(&testing.T{}, e, "ID")
fmt.Println(fld.Tag.Get("json"))
fmt.Println(fld.Tag.Get("validate"))
// Output:
// id
// required
```

`GetField` requires a pointer to struct. Passing a non-pointer or a
pointer to a non-struct calls `t.Error` and returns the zero
`reflect.StructField`.

### GetValue

`GetValue` returns the `reflect.Value` of the named field. It accepts
both pointer and non-pointer structs:

<!-- gmdoceg:ExampleGetValue -->
```go
type Event struct {
    ID    string `json:"id" validate:"required"`
    Score int    `json:"score"`
}
e := &Event{ID: "evt-1"}
val := GetValue(&testing.T{}, e, "ID")
fmt.Println(val.Kind())
fmt.Println(val.String())
// Output:
// string
// evt-1
```

If the named field does not exist, `GetValue` calls `t.Error` and
returns the zero `reflect.Value`.
