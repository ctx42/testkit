// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package oskit

import (
	"os"

	"github.com/ctx42/testing/pkg/must"
)

// Mustwd is a wrapper around [os.Getwd] which panics on error.
func Mustwd() string { return must.Value(os.Getwd()) }
