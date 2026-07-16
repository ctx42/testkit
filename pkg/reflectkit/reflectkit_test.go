// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package reflectkit

import (
	"reflect"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_GetField(t *testing.T) {
	s := &struct {
		F int `url:"f"`
	}{}

	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := GetField(tspy, s, "F")

		// --- Then ---
		assert.Equal(t, "F", have.Name)
	})

	t.Run("error - empty field name", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("the struct field name must not be empty")
		tspy.Close()

		// --- When ---
		have := GetField(tspy, s, "")

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("error - not existing field", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		want := "struct `*struct { F int \"url:\\\"f\\\"\" }` has no field `A`"
		tspy.ExpectLogEqual(want)
		tspy.Close()

		// --- When ---
		have := GetField(tspy, s, "A")

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("error - nil", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("pointer to struct is required, got nil")
		tspy.Close()

		// --- When ---
		have := GetField(tspy, nil, "F")

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("error - not a pointer", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("type `int` is not a pointer to struct")
		tspy.Close()

		// --- When ---
		have := GetField(tspy, 1, "F")

		// --- Then ---
		assert.Zero(t, have)
	})

	t.Run("error - not a struct", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("type `*int` is not struct")
		tspy.Close()

		// --- When ---
		have := GetField(tspy, new(int), "F")

		// --- Then ---
		assert.Zero(t, have)
	})
}

func Test_GetValue(t *testing.T) {
	s := &struct{ F int }{}

	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := GetValue(tspy, s, "F")

		// --- Then ---
		assert.True(t, have.IsValid())
	})

	t.Run("returned value is settable", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()
		sv := &struct{ F int }{}
		fld := GetValue(tspy, sv, "F")

		// --- When ---
		fld.Set(reflect.ValueOf(1))

		// --- Then ---
		assert.Equal(t, 1, sv.F)
	})

	t.Run("error - empty field name", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("the struct field name must not be empty")
		tspy.Close()

		// --- When ---
		have := GetValue(tspy, s, "")

		// --- Then ---
		assert.False(t, have.IsValid())
	})

	t.Run("error - not existing field", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("cannot get value for `*struct { F int }.A`")
		tspy.Close()

		// --- When ---
		have := GetValue(tspy, s, "A")

		// --- Then ---
		assert.False(t, have.IsValid())
	})

	t.Run("error - not a struct", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("type `int` is not struct")
		tspy.Close()

		// --- When ---
		have := GetValue(tspy, 1, "A")

		// --- Then ---
		assert.False(t, have.IsValid())
	})

	t.Run("error - nil", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		want := "struct or pointer to struct is required, got nil"
		tspy.ExpectLogEqual(want)
		tspy.Close()

		// --- When ---
		have := GetValue(tspy, nil, "F")

		// --- Then ---
		assert.False(t, have.IsValid())
	})

	t.Run("error - typed nil pointer", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		want := "cannot get value for `*struct { F int }.F`"
		tspy.ExpectLogEqual(want)
		tspy.Close()
		var sv *struct{ F int }

		// --- When ---
		have := GetValue(tspy, sv, "F")

		// --- Then ---
		assert.False(t, have.IsValid())
	})
}
