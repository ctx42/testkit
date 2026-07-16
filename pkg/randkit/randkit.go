// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package randkit provides random test helpers for generating strings,
// file names, identifiers, and other test fixtures. The default source
// is [math/rand/v2]'s automatically-seeded global PRNG, which is fast
// and sufficiently random for test use. Pass [WithSeed] to switch to a
// deterministic source with stable output.
package randkit

import (
	mrand "math/rand/v2"
	"path/filepath"
	"strings"
)

// Uppercase is the list of uppercase letters.
const Uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Lowercase is the list of lowercase letters.
const Lowercase = "abcdefghijklmnopqrstuvwxyz"

// Letters list of lowercase and uppercase letters.
const Letters = Lowercase + Uppercase

// Digits is a list of digits 0 to 9.
const Digits = "0123456789"

// WithChars is a [Str] option setting the list of characters to use when
// generating random strings. All passed strings are concatenated in the
// order they were passed.
func WithChars(list ...string) func(*options) {
	return func(opts *options) {
		opts.chars = strings.Join(list, "")
	}
}

// WithPrefix is a [Str] option setting prefix for generated strings.
func WithPrefix(prefix string) func(*options) {
	return func(opt *options) { opt.prefix = prefix }
}

// WithSuffix is a [Str] option setting suffix for generated strings.
func WithSuffix(suffix string) func(*options) {
	return func(opt *options) { opt.suffix = suffix }
}

// WithExt is a [Str] option alias for [WithSuffix].
func WithExt(ext string) func(*options) { return WithSuffix(ext) }

// WithLen is a [Str] option setting the length for the generated string.
func WithLen(n int) func(*options) {
	return func(opt *options) { opt.n = n }
}

// WithSeed replaces the global [math/rand/v2] source with a deterministic
// ChaCha8 PRNG seeded by seed. Use only when a test must assert exact
// generated values — never in production code where unpredictability is
// required.
func WithSeed(seed int64) func(*options) {
	return func(opt *options) { opt.rng = seededRand(seed) }
}

// options represents options for generating random strings.
type options struct {
	chars  string
	n      int
	prefix string
	suffix string
	rng    func(n int) int
}

// Str returns a random string based on provided options. When no options
// are given, the generated string will be 10 characters long containing
// only letters.
func Str(opts ...func(*options)) string {
	def := options{
		chars: Letters,
		n:     10,
		rng:   globalRand,
	}
	for _, opt := range opts {
		opt(&def)
	}
	return def.prefix + randStr(def.chars, def.n, def.rng) + def.suffix
}

// FileName returns a random file name. By default, the file name is 7
// letters long [a-zA-Z] with the prefix "file-" and extension ".txt".
func FileName(dir string, opts ...func(*options)) string {
	def := options{
		chars:  Letters,
		n:      7,
		prefix: "file-",
		suffix: ".txt",
		rng:    globalRand,
	}
	for _, opt := range opts {
		opt(&def)
	}
	name := def.prefix + randStr(def.chars, def.n, def.rng) + def.suffix
	return filepath.Join(dir, name)
}

// Int generates a random integer in the range [1, max].
func Int(maximum int, opts ...func(*options)) int {
	def := options{rng: globalRand}
	for _, opt := range opts {
		opt(&def)
	}
	return def.rng(maximum) + 1
}

// Password returns an n-character random password drawn from letters
// [a-zA-Z] and digits [0-9]. No special characters are included.
func Password(n int, opts ...func(*options)) string {
	return Str(append([]func(*options){
		WithChars(Letters, Digits),
		WithLen(n),
	}, opts...)...)
}

// randStr returns n random characters drawn from chars.
func randStr(chars string, n int, rng func(int) int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rng(len(chars))]
	}
	return string(b)
}

// globalRand returns a uniform random int in [0, n) using the global
// [math/rand/v2] source, which is automatically seeded at program start.
func globalRand(n int) int {
	return mrand.IntN(n)
}

// seededRand returns an intn function backed by a deterministic ChaCha8
// PRNG seeded by seed. The returned function panics if n <= 0.
// Intended for tests that need stable, reproducible output only.
func seededRand(seed int64) func(n int) int {
	src := mrand.NewChaCha8(seedToBytes(seed))
	r := mrand.New(src)
	return func(n int) int {
		if n <= 0 {
			panic("invalid argument to RandIntn: n must be > 0")
		}
		return r.IntN(n)
	}
}

// seedToBytes converts an int64 seed into the [32]byte key required by
// ChaCha8. The seed is encoded little-endian in the first 8 bytes; the
// remaining 24 bytes are derived by XOR-ing the corresponding seed byte
// with the byte's index, spreading entropy across the full key.
func seedToBytes(seed int64) [32]byte {
	var b [32]byte
	for i := range 8 {
		b[i] = byte(seed >> (i * 8)) //nolint:gosec
	}
	for i := 8; i < 32; i++ {
		b[i] = byte(seed>>((i%8)*8)) ^ byte(i) //nolint:gosec
	}
	return b
}
