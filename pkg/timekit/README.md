<!-- TOC -->
* [The `timekit` package](#the-timekit-package)
  * [Clocks](#clocks)
    * [ClockFixed](#clockfixed)
    * [ClockDeterministic](#clockdeterministic)
    * [TikTak](#tiktak)
    * [ClockStartingAt](#clockstartingat)
<!-- TOC -->

# The `timekit` package

Production code that reads the wall clock directly with `time.Now` is
hard to test reliably: results drift from one run to the next, and
time-dependent edge cases are awkward to reproduce. The fix is to
inject a clock — a `func() time.Time` — and swap in a controlled one
during tests.

Every `timekit` constructor returns a `func() time.Time`, the same
signature as `time.Now`, so it drops in wherever production code
accepts a clock dependency:

- `ClockFixed` — always returns the same instant.
- `ClockDeterministic` — advances by a fixed duration on each call.
- `TikTak` — `ClockDeterministic` with a one-second tick.
- `ClockStartingAt` — advances in real time from a chosen instant.

## Clocks

### ClockFixed

`ClockFixed` freezes time: every call returns the same instant, no
matter how many times the clock is read. Use it when the test must not
observe any passage of time.

<!-- gmdoceg:ExampleClockFixed -->
```go
clk := timekit.ClockFixed(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))

fmt.Println(clk())
fmt.Println(clk())
fmt.Println(clk())
// Output:
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 00:00:00 +0000 UTC
```

### ClockDeterministic

`ClockDeterministic` advances by a fixed tick on every call, starting
from the given instant. It is ideal for asserting sequence-dependent
time logic without sleeping.

<!-- gmdoceg:ExampleClockDeterministic -->
```go
start := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
clk := timekit.ClockDeterministic(start, time.Hour)

fmt.Println(clk())
fmt.Println(clk())
fmt.Println(clk())
// Output:
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 01:00:00 +0000 UTC
// 2022-01-01 02:00:00 +0000 UTC
```

### TikTak

`TikTak` is `ClockDeterministic` with a fixed one-second tick — the
common case, spelled out so you do not have to pass the duration.

<!-- gmdoceg:ExampleTikTak -->
```go
start := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
clk := timekit.TikTak(start)

fmt.Println(clk())
fmt.Println(clk())
fmt.Println(clk())
// Output:
// 2022-01-01 00:00:00 +0000 UTC
// 2022-01-01 00:00:01 +0000 UTC
// 2022-01-01 00:00:02 +0000 UTC
```

### ClockStartingAt

`ClockStartingAt` returns a clock that advances in real wall-clock
time but is offset so that its first reading is the given instant.
Use it when a test needs plausibly-moving timestamps anchored to a
specific era. Because it tracks real time, its output is not
reproducible and so has no runnable example:

```go
// Tests see timestamps that start at the chosen instant and then
// advance at real speed.
clk := timekit.ClockStartingAt(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))
sched := NewScheduler(clk)
```
