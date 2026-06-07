<!-- TOC -->
* [The `timekit` package](#the-timekit-package)
  * [Clocks](#clocks)
    * [ClockFixed](#clockfixed)
    * [ClockDeterministic](#clockdeterministic)
    * [TikTak](#tiktak)
<!-- TOC -->

# The `timekit` package

The `timekit` package provides `time.Time` related helpers.

## Clocks

All clock functions return `func() time.Time`, the same signature as
`time.Now`, so they can be injected wherever the production code accepts
a clock dependency.

- `ClockStartingAt` — returns real wall time offset to start at a given
  instant.
- `ClockFixed` — always returns the same instant.
- `ClockDeterministic` — advances by a fixed duration on each call.
- `TikTak` — like `ClockDeterministic` with a one-second tick.

### ClockFixed

<!-- gmdoceg:ExampleClockFixed -->
```go
clk := ClockFixed(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))

fmt.Println(clk())
fmt.Println(clk())
fmt.Println(clk())
// Output:
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 00:00:00 +0000 UTC
```

### ClockDeterministic

<!-- gmdoceg:ExampleClockDeterministic -->
```go
start := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
clk := ClockDeterministic(start, time.Hour)

fmt.Println(clk())
fmt.Println(clk())
fmt.Println(clk())
// Output:
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 01:00:00 +0000 UTC
// 2022-01-01 02:00:00 +0000 UTC
```

### TikTak

<!-- gmdoceg:ExampleTikTak -->
```go
start := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
clk := TikTak(start)

fmt.Println(clk())
fmt.Println(clk())
fmt.Println(clk())
// Output:
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 00:00:01 +0000 UTC
// 2022-01-01 00:00:02 +0000 UTC
```
