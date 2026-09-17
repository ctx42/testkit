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
// the file at pth to be created then reads it and returns its content. Returns
// empty string and error if the file cannot be read or timeout is triggered.
// The throttle defaults to 50ms; pass [check.Option] values in opts to
// override it (e.g. [check.WithWaitThrottle]).
//
// Writers often create the file before writing to it, so a file which exists
// but is still empty is re-read (with the same timeout and throttle) until it
// has content. An empty string with a nil error is returned when the file is
// still empty when the timeout is triggered.
//
// Panics when timeout is an invalid [time.Duration] string.
func Wait4File(timeout, pth string, opts ...any) (string, error) {
	fn := func() bool { _, err := os.Stat(pth); return err == nil }
	opts = append([]any{check.WithWaitThrottle(50 * time.Millisecond)}, opts...)
	if err := check.Wait(timeout, fn, opts...); err != nil {
		err = notice.From(err).
			SetHeader("timeout waiting for file read").
			Append("file", "%s", pth)
		return "", err
	}

	var data []byte
	var rdrErr error
	read := func() bool {
		data, rdrErr = os.ReadFile(pth) //nolint:gosec
		return rdrErr != nil || len(data) > 0
	}
	// Timeout is not an error here — the file may be legitimately empty.
	_ = check.Wait(timeout, read, opts...)
	if rdrErr != nil {
		err := notice.From(rdrErr).
			SetHeader("reading file").
			Append("file", "%s", pth)
		return "", err
	}
	return string(data), nil
}
