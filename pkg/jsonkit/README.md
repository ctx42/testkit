# The `jsonkit` package

`jsonkit` provides JSON helpers for tests. The helpers free tests from
`if err != nil` boilerplate when arranging fixtures and asserting results:
instead of returning an error, each helper takes a `tester.T` and marks the
test as failed when the underlying JSON operation fails.

## Import

```go
import "github.com/ctx42/testkit/pkg/jsonkit"
```

## Usage

```go
func Test_Example(t *testing.T) {
	data := jsonkit.To(t, map[string]any{"name": "Bob", "age": 42})

	data, age := jsonkit.DeleteKey(t, data, "age")

	assert.Equal(t, 42.0, age)
	assert.JSON(t, `{"name": "Bob"}`, string(data))
}
```
