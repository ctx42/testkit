// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package httpkit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_RespWriter_Header(t *testing.T) {
	t.Run("headers not yet written", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		rec := httptest.NewRecorder()
		rec.Header().Set("A", "1")
		rsw := NewRespWriter(tspy, rec)

		// --- When ---
		have := rsw.Header()

		// --- Then ---
		assert.Equal(t, "1", have.Get("A"))
	})

	t.Run("error - headers already written", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "headers already written:\n" +
			"  last status: 200"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		rec := httptest.NewRecorder()
		rec.Header().Set("A", "1")
		rsw := NewRespWriter(tspy, rec)
		_, _ = rsw.Write([]byte{1, 2, 3})

		// --- When ---
		have := rsw.Header()

		// --- Then ---
		assert.Equal(t, "1", have.Get("A"))
	})
}

func Test_RespWriter_Write(t *testing.T) {
	// --- Given ---
	tspy := tester.New(t)
	tspy.Close()

	rec := httptest.NewRecorder()
	rsw := NewRespWriter(tspy, rec)

	// --- When ---
	have, err := rsw.Write([]byte{1, 2, 3})

	// --- Then ---
	assert.NoError(t, err)
	assert.Equal(t, 3, have)
	assert.Equal(t, "\x01\x02\x03", rec.Body.String())
	assert.True(t, rsw.written)
}

func Test_RespWriter_WriteHeader(t *testing.T) {
	t.Run("headers not yet written", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		rec := httptest.NewRecorder()
		rec.Header().Set("A", "1")
		rsw := NewRespWriter(tspy, rec)

		// --- When ---
		rsw.WriteHeader(404)

		// --- Then ---
		assert.Equal(t, "1", rec.Header().Get("A"))
		assert.Equal(t, 404, rec.Code)
	})

	t.Run("error - headers already written", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "headers already written:\n" +
			"  last status: 200\n" +
			"   new status: 404"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		rec := httptest.NewRecorder()
		rec.Header().Set("A", "1")
		rsw := NewRespWriter(tspy, rec)
		_, _ = rsw.Write([]byte{1, 2, 3})

		// --- When ---
		rsw.WriteHeader(404)

		// --- Then ---
		assert.Equal(t, "1", rec.Header().Get("A"))
		assert.Equal(t, 200, rec.Code)
	})
}

func Test_RespWriterMW(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		h0 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("h0 "))
		})

		// --- When ---
		have := RespWriterMW(tspy, h0)

		// --- Then ---
		rec := httptest.NewRecorder()
		have.ServeHTTP(rec, nil)
		assert.Equal(t, "h0 ", rec.Body.String())
	})

	t.Run("error - setting status twice", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "headers already written:\n" +
			"  last status: 400\n" +
			"   new status: 200"
		tspy.ExpectLogEqual(wMsg)
		tspy.Close()

		h0 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(400)
			w.WriteHeader(200)
			_, _ = w.Write([]byte("h0 "))
		})

		// --- When ---
		have := RespWriterMW(tspy, h0)

		// --- Then ---
		rec := httptest.NewRecorder()
		have.ServeHTTP(rec, nil)
		assert.Equal(t, "h0 ", rec.Body.String())
	})
}
