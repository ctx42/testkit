// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package timekit provides controllable and deterministic clocks for testing
// code that depends on time.
//
// It offers factory functions that return values with the same signature as
// [time.Now]. These are essential when testing time-sensitive logic without
// relying on the real system clock.
//
// The helpers integrate naturally with [tester.T] and the assertion packages.
//
// See the package README for usage patterns and motivation. See
// examples_test.go for executable demonstrations.
//
// Key entry points:
//   - [ClockStartingAt] — returns time advancing from a given offset
//   - [ClockFixed] — always returns the same instant
//   - [ClockDeterministic] — advances by a fixed tick on every call
//   - [TikTak] — convenience for one-second deterministic ticks
package timekit

import (
	"sync"
	"time"
)

// ClockStartingAt returns a clock that behaves as if "now" was set to the
// given instant at the moment of creation. Subsequent calls advance in real
// time from that base.
func ClockStartingAt(tim time.Time) func() time.Time {
	now := time.Now()
	return func() time.Time {
		return tim.Add(time.Since(now))
	}
}

// ClockFixed returns a clock that always reports the exact same instant,
// regardless of how much real time passes. Useful for freezing time in tests.
func ClockFixed(tim time.Time) func() time.Time {
	return func() time.Time {
		return tim
	}
}

// ClockDeterministic returns a clock that advances by exactly "tick" on every
// invocation, independent of real elapsed time. Ideal for fully deterministic
// time progression in tests.
func ClockDeterministic(start time.Time, tick time.Duration) func() time.Time {
	now := start.Add(-tick)
	var guard sync.Mutex
	return func() time.Time {
		guard.Lock()
		defer guard.Unlock()
		now = now.Add(tick)
		return now
	}
}

// TikTak is a convenience wrapper around [ClockDeterministic] that advances
// exactly one second on every call.
func TikTak(start time.Time) func() time.Time {
	return ClockDeterministic(start, time.Second)
}
