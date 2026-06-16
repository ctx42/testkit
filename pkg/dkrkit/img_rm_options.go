// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"time"
)

// ImgRmOption represents an option for [Docker.ImgRm] and [DockerT.ImgRm].
type ImgRmOption func(*ImgRmOptions)

// ImgRmOptions represents Docker image remove options.
type ImgRmOptions struct {
	// If the container based on the image is running sleep for a given time
	// between retries.
	sleep time.Duration

	// Maximum number of retries.
	tries int

	// When set, DockerT.ImgRm logs errors via t.Log instead of failing the
	// test via t.Error. Has no effect on Docker.ImgRm, which always returns
	// errors to the caller.
	ignoreErrors bool
}

// DefaultImgRmOptions returns default image remove options with any provided
// options applied.
func DefaultImgRmOptions(ops ...ImgRmOption) *ImgRmOptions {
	opts := &ImgRmOptions{
		sleep: 250 * time.Millisecond,
		tries: 20,
	}
	for _, op := range ops {
		op(opts)
	}
	return opts
}

// WithImgRmSleep is an option for setting sleep duration between image removal
// retries.
func WithImgRmSleep(d time.Duration) ImgRmOption {
	return func(opts *ImgRmOptions) { opts.sleep = d }
}

// WithImgRmTries is an option for setting the number of image removal retries.
func WithImgRmTries(cnt int) ImgRmOption {
	return func(opts *ImgRmOptions) { opts.tries = cnt }
}

// WithImgRmIgnoreErrors is a [DockerT.ImgRm] option that downgrades errors
// from test failures to log messages. Instead of calling t.Error, the method
// calls t.Log and returns false, so the test continues running.
//
// It has no effect on [Docker.ImgRm], which always returns errors to the
// caller regardless of this option.
func WithImgRmIgnoreErrors() ImgRmOption {
	return func(iro *ImgRmOptions) { iro.ignoreErrors = true }
}
