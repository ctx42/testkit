// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package iokit

import (
	"io"
)

// ErrorReadCloser implements [io.ReadCloser] by embedding an [ErrorReader]
// and adding Close behavior controllable via options.
//
// See [ErrReadCloser] for the constructor.
type ErrorReadCloser struct {
	*ErrorReader
	cls io.Closer
}

var _ io.ReadCloser = (*ErrorReadCloser)(nil)

// ErrReadCloser wraps "src" and allows up to n bytes to be read before
// returning an error. If n < 0 it behaves normally.
//
// Use With*Err options to customize errors (original Close is still called).
func ErrReadCloser(src io.ReadCloser, n int, opts ...Option) *ErrorReadCloser {
	return &ErrorReadCloser{
		ErrorReader: ErrReader(src, n, opts...),
		cls:         src,
	}
}

// implements [io.Closer]. The underlying Close is always called; a custom
// error set via [WithCloseErr] overrides its result.
func (rc *ErrorReadCloser) Close() error {
	err := rc.cls.Close() // The underlying Close method is always called.
	if rc.errClose != nil {
		return rc.errClose
	}
	return err
}
