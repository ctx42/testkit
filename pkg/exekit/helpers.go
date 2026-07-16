// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package exekit

import (
	"strings"
)

// envSplit parses environment (in [os.Environ] format) and returns it as key
// value map.
func envSplit(env []string) map[string]string {
	out := map[string]string{}
	for _, s := range env {
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "=", 2)
		switch len(parts) {
		case 1:
			out[s] = ""
		case 2:
			out[parts[0]] = parts[1]
		}
	}
	return out
}

// envJoin gets environment variables as a map of key values and returns it as
// a slice. Which is the same return format as os.Environ function.
func envJoin(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}
