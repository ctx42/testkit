// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dkrkit

import (
	"maps"
	"sync"
	"time"
)

// CtrRunOption is an option for [Docker.CtrRun] and [DockerT.CtrRun].
type CtrRunOption func(*CtrRunOptions)

// CtrRunOptions represents Docker container run options.
type CtrRunOptions struct {
	// Extra arguments to add to the end of "docker run" arguments.
	args []string

	// Labels to set on the container.
	labels map[string]string

	// Absolute path to the file where container ID should be written to. When
	// empty, the "--cidfile" option will not be used.
	cidPth string

	// Channel the container ID will be sent on. Closed exactly once via
	// cidClose — either after the ID is delivered or on timeout.
	cidCh chan string

	// Closes cidCh exactly once; safe to call multiple times.
	cidClose func()

	// Polling interval used when waiting for the CID file to appear.
	cidPoll time.Duration

	// Run container in the background.
	detach bool

	// Control whether to remove a container after it exits.
	remove bool
}

// DefaultCtrRunOptions returns default container run options with any provided
// options applied.
func DefaultCtrRunOptions(ops ...CtrRunOption) *CtrRunOptions {
	def := &CtrRunOptions{
		cidPoll: 50 * time.Millisecond,
		remove:  true,
	}
	for _, op := range ops {
		op(def)
	}
	return def
}

// WithCtrRunArgs is an option for appending run extra args.
func WithCtrRunArgs(args ...string) CtrRunOption {
	return func(opts *CtrRunOptions) { opts.args = append(opts.args, args...) }
}

// WithCtrRunLabel is an option for setting container label.
func WithCtrRunLabel(name, value string) CtrRunOption {
	return func(opts *CtrRunOptions) {
		if opts.labels == nil {
			opts.labels = make(map[string]string)
		}
		opts.labels[name] = value
	}
}

// WithCtrRunLabels is an option for setting container labels.
func WithCtrRunLabels(labels map[string]string) CtrRunOption {
	return func(opts *CtrRunOptions) {
		if opts.labels == nil {
			opts.labels = make(map[string]string)
		}
		maps.Copy(opts.labels, labels)
	}
}

// WithCtrRunCIDPth is an option for setting the path to the
// CID file.
func WithCtrRunCIDPth(pth string) CtrRunOption {
	return func(opts *CtrRunOptions) { opts.cidPth = pth }
}

// WithCtrRunCID returns a channel on which the container ID will be sent once
// the container starts, and the option to enable it. The channel is closed
// exactly once — either after the ID is delivered or on timeout.
//
// [WithCtrRunCIDPth] may be combined with this option to control where the
// temporary cidfile is written; without it a temp directory is created and
// cleaned up automatically.
func WithCtrRunCID() (<-chan string, CtrRunOption) {
	ch := make(chan string, 1)
	var once sync.Once
	fn := func(opts *CtrRunOptions) {
		opts.cidCh = ch
		opts.cidClose = func() { once.Do(func() { close(ch) }) }
	}
	return ch, fn
}

// WithCtrRunDetach is an option for setting "--detach" option.
func WithCtrRunDetach() CtrRunOption {
	return func(opts *CtrRunOptions) { opts.detach = true }
}

// WithCtrRunNoRemove is an option for preventing removal of the
// docker container after it exits.
func WithCtrRunNoRemove() CtrRunOption {
	return func(opts *CtrRunOptions) { opts.remove = false }
}
