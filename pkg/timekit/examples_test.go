// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package timekit_test

import (
	"fmt"
	"time"

	"github.com/ctx42/testkit/pkg/timekit"
)

func ExampleClockFixed() {
	clk := timekit.ClockFixed(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC))

	fmt.Println(clk())
	fmt.Println(clk())
	fmt.Println(clk())
	// Output:
	// 2022-01-01 00:00:00 +0000 UTC
	// 2022-01-01 00:00:00 +0000 UTC
	// 2022-01-01 00:00:00 +0000 UTC
}

func ExampleClockDeterministic() {
	start := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := timekit.ClockDeterministic(start, time.Hour)

	fmt.Println(clk())
	fmt.Println(clk())
	fmt.Println(clk())
	// Output:
	// 2022-01-01 00:00:00 +0000 UTC
	// 2022-01-01 01:00:00 +0000 UTC
	// 2022-01-01 02:00:00 +0000 UTC
}

func ExampleTikTak() {
	start := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	clk := timekit.TikTak(start)

	fmt.Println(clk())
	fmt.Println(clk())
	fmt.Println(clk())
	// Output:
	// 2022-01-01 00:00:00 +0000 UTC
	// 2022-01-01 00:00:01 +0000 UTC
	// 2022-01-01 00:00:02 +0000 UTC
}
