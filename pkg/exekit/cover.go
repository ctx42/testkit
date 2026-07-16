// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package exekit

import (
	"strings"
)

// IsWithCoverage reports whether tests were run with coverage enabled by
// inspecting program arguments (usually [os.Args]). Detection is based on
// the -coverprofile and -test.coverprofile flag forms only. As the first
// return value, it returns the coverage profile path as provided.
func IsWithCoverage(args []string) (string, bool) {
	for i, arg := range args {
		part := "-coverprofile="
		if strings.HasPrefix(arg, part) {
			return arg[len(part):], true
		}

		part = "-coverprofile"
		if strings.HasPrefix(arg, part) {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", false
		}

		part = "-test.coverprofile="
		if strings.HasPrefix(arg, part) {
			return arg[len(part):], true
		}

		part = "-test.coverprofile"
		if strings.HasPrefix(arg, part) {
			if i+1 < len(args) {
				return args[i+1], true
			}
			return "", false
		}
	}
	return "", false
}

// MaybeAddGoCovDir using [IsWithCoverage] adds, if needed, the GOCOVERDIR
// variable to the passed "env" and returns it. The getDir function is called
// to get the path to use as the GOCOVERDIR value. If GOCOVERDIR is already
// present in env it will not be overridden.
func MaybeAddGoCovDir(env, args []string, getDir func() string) []string {
	if _, ok := IsWithCoverage(args); !ok {
		return env
	}
	for _, kv := range env {
		if strings.HasPrefix(kv, "GOCOVERDIR=") {
			return env
		}
	}
	return append(env, "GOCOVERDIR="+getDir())
}
