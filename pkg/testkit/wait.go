// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package testkit

import (
	"os"
	"time"

	"github.com/ctx42/testing/pkg/check"
	"github.com/ctx42/testing/pkg/notice"
)

// Wait4File waits with timeout (string representation of [time.Duration]) for
// the file at pth to be created than reads it and returns its content. Returns
// empty string and error if the file cannot be read or timeout is triggered.
// The throttle defaults to 50ms; pass [check.Option] values in opts to
// override it (e.g. [check.WithWaitThrottle]).
//
// Panics when timeout is an invalid [time.Duration] string.
func Wait4File(timeout, pth string, opts ...any) (string, error) {
	fn := func() bool { _, err := os.Stat(pth); return err == nil }
	throttle := 50 * time.Millisecond
	opts = append([]any{check.WithWaitThrottle(throttle)}, opts...)
	err := check.Wait(timeout, fn, opts...)
	if err != nil {
		err = notice.From(err).
			SetHeader("timeout waiting for file read").
			Append("file", "%s", pth)
		return "", err
	}
	var reads int
	for {
		data, err := os.ReadFile(pth) // nolint:gosec
		if err != nil {
			return "", err
		}
		reads++
		// There is an edge case where the file already exists, but
		// it was not yet written to. We try again when read data is empty.
		if len(data) == 0 && reads == 1 {
			time.Sleep(throttle)
			continue
		}
		return string(data), nil
	}
}
