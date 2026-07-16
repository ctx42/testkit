// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package reflectkit provides lightweight reflection utilities for testing.
package reflectkit

import (
	"reflect"

	"github.com/ctx42/testing/pkg/tester"
)

// GetField returns the [reflect.StructField] for the named field in struct s.
// The argument s must be a pointer to struct. On any error it reports via
// t.Error (never panics) and returns the zero StructField.
func GetField(t tester.T, s any, name string) reflect.StructField {
	t.Helper()

	if name == "" {
		t.Error("the struct field name must not be empty")
		return reflect.StructField{}
	}
	if s == nil {
		t.Error("pointer to struct is required, got nil")
		return reflect.StructField{}
	}
	typ := reflect.TypeOf(s)
	if typ.Kind() != reflect.Pointer {
		t.Errorf("type `%T` is not a pointer to struct", s)
		return reflect.StructField{}
	}
	typ = typ.Elem()
	if typ.Kind() != reflect.Struct {
		t.Errorf("type `%T` is not struct", s)
		return reflect.StructField{}
	}
	fld, exist := typ.FieldByName(name)
	if !exist {
		t.Errorf("struct `%T` has no field %#q", s, name)
		return reflect.StructField{}
	}
	return fld
}

// GetValue returns the [reflect.Value] for the named field in struct s.
// The argument s may be a struct or pointer to struct. On any error it
// reports via t.Error (never panics) and returns the zero Value.
func GetValue(t tester.T, s any, name string) reflect.Value {
	t.Helper()

	if name == "" {
		t.Error("the struct field name must not be empty")
		return reflect.Value{}
	}
	if s == nil {
		t.Error("struct or pointer to struct is required, got nil")
		return reflect.Value{}
	}

	typ := reflect.TypeOf(s)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		t.Errorf("type `%T` is not struct", s)
		return reflect.Value{}
	}

	sv := reflect.ValueOf(s)
	iv := reflect.Indirect(sv)
	if !iv.IsValid() {
		t.Errorf("cannot get value for `%T.%s`", s, name)
		return reflect.Value{}
	}
	fld := iv.FieldByName(name)
	if !fld.IsValid() {
		t.Errorf("cannot get value for `%T.%s`", s, name)
		return reflect.Value{}
	}
	return fld
}
