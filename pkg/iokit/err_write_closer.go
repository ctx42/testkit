// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package iokit

import (
	"io"
)

// ErrorWriteCloser implements [io.WriteCloser] by embedding an [ErrorWriter]
// and adding controllable Close behavior.
//
// See [ErrWriteCloser] for the constructor.
type ErrorWriteCloser struct {
	*ErrorWriter
	cls io.Closer
}

var _ io.WriteCloser = (*ErrorWriteCloser)(nil)

// ErrWriteCloser wraps "dst" and allows up to n bytes to be written before
// returning an error. If n < 0 it behaves normally.
//
// Use With*Err options to customize write and close errors (original Close
// is still called).
func ErrWriteCloser(
	dst io.WriteCloser,
	n int,
	opts ...Option,
) *ErrorWriteCloser {
	return &ErrorWriteCloser{
		ErrorWriter: ErrWriter(dst, n, opts...),
		cls:         dst,
	}
}

// implements [io.Closer]. The underlying Close is always called; a custom
// error set via [WithCloseErr] overrides its result.
func (wc *ErrorWriteCloser) Close() error {
	err := wc.cls.Close() // The underlying Close method is always called.
	if wc.errClose != nil {
		return wc.errClose
	}
	return err
}
