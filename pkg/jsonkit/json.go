// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package jsonkit provides JSON helpers for tests.
//
// Each helper takes a [tester.T] and, instead of returning an error, marks the
// test as failed when the underlying [encoding/json] operation fails: [To] and
// [ToReader] marshal a value, [ToMap], [From], and [FromReader] unmarshal it,
// and [DeleteKey] removes a key from a JSON object.
package jsonkit

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// To is a wrapper for [json.Marshal]. On error, it marks the test as failed
// and returns nil.
func To(t tester.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		msg := notice.New("expected no error marshalling value").
			Append("error", "%s", err)
		t.Error(msg)
		return nil
	}
	return data
}

// ToReader marshals v to JSON and returns an [io.Reader] for it. On error, it
// marks the test as failed and returns a reader over no bytes.
func ToReader(t tester.T, v any) io.Reader {
	t.Helper()
	return bytes.NewReader(To(t, v))
}

// ToMap unmarshals JSON to map[string]any and returns it on success. On
// failure, it marks the test as failed and returns nil.
func ToMap(t tester.T, data []byte) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		msg := notice.New("expected no error unmarshalling data as map").
			Append("error", "%s", err)
		t.Error(msg)
		return nil
	}
	return m
}

// From is a wrapper around [json.Unmarshal] returning true on success. On
// error, it marks the test as failed and returns false.
func From(t tester.T, data []byte, v any) bool {
	t.Helper()
	if err := json.Unmarshal(data, v); err != nil {
		msg := notice.New("expected no error unmarshalling data").
			Append("error", "%s", err)
		t.Error(msg)
		return false
	}
	return true
}

// FromReader reads JSON from r and unmarshals it to v. On success it returns
// true. On error, it marks the test as failed and returns false.
func FromReader(t tester.T, r io.Reader, v any) bool {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		msg := notice.New("expected no error decoding JSON").
			Append("error", "%s", err)
		t.Error(msg)
		return false
	}
	return true
}

// DeleteKey expects data to be a valid JSON object with the given key, the
// JSON is unmarshaled to `map[string]any` type then the key is deleted from
// the map, and it is marshaled again and returned with the value of the
// deleted key as the second argument. If data is not valid JSON, it cannot be
// unmarshalled to the map, or the key does not exist, it marks the test as
// failed with the appropriate error message. On error the passed data slice is
// returned unchanged; otherwise the re-marshaled JSON without the key.
func DeleteKey(t tester.T, data []byte, key string) ([]byte, any) {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		msg := notice.New("expected no error unmarshalling data as map").
			Append("map type", "%T", m).
			Append("error", "%s", err)
		t.Error(msg)
		return data, nil
	}
	if val, ok := m[key]; ok {
		delete(m, key)
		data, _ = json.Marshal(m) //nolint:errchkjson
		return data, val
	}

	msg := notice.New("expected the map to have a key").Append("key", "%s", key)
	t.Error(msg)
	return data, nil
}
