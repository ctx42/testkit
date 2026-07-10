// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package jsonkit

import (
	"io"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_To(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		s := struct {
			FStr string `json:"f_str"`
		}{
			FStr: "abc",
		}

		// --- When ---
		have := To(tspy, s)

		// --- Then ---
		assert.JSON(t, `{"f_str": "abc"}`, string(have))
	})

	t.Run("nil value", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := To(tspy, nil)

		// --- Then ---
		assert.Equal(t, "null", string(have))
	})

	t.Run("error - unsupported type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "" +
			"expected no error marshalling value:\n" +
			"  error: json: unsupported type: func()"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		s := struct {
			FFn func() `json:"fn"`
		}{
			FFn: func() {},
		}

		// --- When ---
		have := To(tspy, s)

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_ToReader(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		s := struct {
			FStr string `json:"f_str"`
		}{
			FStr: "abc",
		}

		// --- When ---
		have := ToReader(tspy, s)

		// --- Then ---
		data := must.Value(io.ReadAll(have))
		assert.JSON(t, `{"f_str": "abc"}`, string(data))
	})

	t.Run("error - unsupported type", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "" +
			"expected no error marshalling value:\n" +
			"  error: json: unsupported type: func()"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		s := struct {
			FFn func() `json:"fn"`
		}{
			FFn: func() {},
		}

		// --- When ---
		have := ToReader(tspy, s)

		// --- Then ---
		data := must.Value(io.ReadAll(have))
		assert.Empty(t, data)
	})
}

func Test_ToMap(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		data := []byte(`{"f_str": "abc"}`)

		// --- When ---
		have := ToMap(tspy, data)

		// --- Then ---
		assert.Equal(t, map[string]any{"f_str": "abc"}, have)
	})

	t.Run("error - invalid json", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("invalid character")
		tspy.Close()

		data := []byte("{!!!}")

		// --- When ---
		have := ToMap(tspy, data)

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("error - not a json object", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectLogContain("cannot unmarshal array")
		tspy.Close()

		data := []byte("[1, 2]")

		// --- When ---
		have := ToMap(tspy, data)

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_From(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		data := []byte(`{"f_str": "abc"}`)

		s := struct {
			FStr string `json:"f_str"`
		}{}

		// --- When ---
		have := From(tspy, data, &s)

		// --- Then ---
		assert.True(t, have)
		assert.Equal(t, "abc", s.FStr)
	})

	t.Run("error - invalid json", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "" +
			"expected no error unmarshalling data:\n" +
			"  error: invalid character '!' " +
			"looking for beginning of object key string"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		data := []byte(`{!!!}`)

		s := struct {
			FStr string `json:"f_str"`
		}{}

		// --- When ---
		have := From(tspy, data, &s)

		// --- Then ---
		assert.False(t, have)
		assert.Equal(t, "", s.FStr)
	})
}

func Test_FromReader(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		data := strings.NewReader(`{"f_str": "abc"}`)

		s := struct {
			FStr string `json:"f_str"`
		}{}

		// --- When ---
		have := FromReader(tspy, data, &s)

		// --- Then ---
		assert.True(t, have)
		assert.Equal(t, "abc", s.FStr)
	})

	t.Run("error - invalid json", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		wMsg := "" +
			"expected no error decoding JSON:\n" +
			"  error: invalid character '!' " +
			"looking for beginning of object key string"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		data := strings.NewReader("{!!!}")

		s := struct {
			FStr string `json:"f_str"`
		}{}

		// --- When ---
		have := FromReader(tspy, data, &s)

		// --- Then ---
		assert.False(t, have)
		assert.Equal(t, "", s.FStr)
	})
}

func Test_DeleteKey(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		data := []byte(`{"A": 1, "B": 2}`)

		// --- When ---
		have, val := DeleteKey(tspy, data, "A")

		// --- Then ---
		assert.Equal(t, `{"B":2}`, string(have))
		assert.Equal(t, 1.0, val)
	})

	t.Run("error - key not found", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("expected the map to have a key:\n  key: C")
		tspy.Close()

		data := []byte(`{"A": 1, "B": 2}`)

		// --- When ---
		have, val := DeleteKey(tspy, data, "C")

		// --- Then ---
		assert.Equal(t, `{"A": 1, "B": 2}`, string(have))
		assert.Nil(t, val)
	})

	t.Run("error - invalid json", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "" +
			"expected no error unmarshalling data as map:\n" +
			"  map type: map[string]interface {}\n" +
			"     error: invalid character"
		tspy.ExpectLogContain(wMsg)
		tspy.Close()

		data := []byte(`{!!!}`)

		// --- When ---
		have, val := DeleteKey(tspy, data, "C")

		// --- Then ---
		assert.Equal(t, `{!!!}`, string(have))
		assert.Nil(t, val)
	})
}
