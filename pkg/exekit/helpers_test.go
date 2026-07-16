// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package exekit

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_envSplit_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env  []string
		want map[string]string
	}{
		{"1", []string{}, map[string]string{}},
		{"1a", []string{""}, map[string]string{}},
		{"2", []string{"A=B"}, map[string]string{"A": "B"}},
		{"3", []string{"A=B=C"}, map[string]string{"A": "B=C"}},
		{"4", []string{"A"}, map[string]string{"A": ""}},
		{"4a", []string{"A="}, map[string]string{"A": ""}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			env := envSplit(tc.env)

			// --- Then ---
			assert.Equal(t, tc.want, env)
		})
	}
}

func Test_envJoin_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env  map[string]string
		want []string
	}{
		{"1", map[string]string{}, []string{}},
		{"1a", map[string]string{}, []string{}},
		{"2", map[string]string{"A": "B"}, []string{"A=B"}},
		{"3", map[string]string{"A": "B=C"}, []string{"A=B=C"}},
		{"4", map[string]string{"A": ""}, []string{"A="}},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			env := envJoin(tc.env)

			// --- Then ---
			assert.Equal(t, tc.want, env)
		})
	}
}
