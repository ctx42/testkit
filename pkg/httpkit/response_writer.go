// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package httpkit

import (
	"net/http"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// RespWriter wraps [http.ResponseWriter] to track calls to its methods
// and report errors if they were misused.
type RespWriter struct {
	http.ResponseWriter          // Wrapped interface.
	status              int      // The latest status.
	written             bool     // True if headers already written.
	t                   tester.T // Test manager.
}

// NewRespWriter returns a new instance of [RespWriter].
func NewRespWriter(t tester.T, w http.ResponseWriter) *RespWriter {
	t.Helper()
	return &RespWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
		t:              t,
	}
}

// Header — triggers a test error if headers were already sent.
func (crw *RespWriter) Header() http.Header {
	if crw.written {
		crw.t.Helper()
		msg := notice.New("headers already written").
			Append("last status", "%d", crw.status)
		crw.t.Error(msg)
	}
	return crw.ResponseWriter.Header()
}

func (crw *RespWriter) Write(data []byte) (int, error) {
	crw.written = true
	return crw.ResponseWriter.Write(data)
}

// WriteHeader — triggers a test error if headers were already sent.
func (crw *RespWriter) WriteHeader(status int) {
	if crw.written {
		crw.t.Helper()
		msg := notice.New("headers already written").
			Append("last status", "%d", crw.status).
			Append("new status", "%d", status)
		crw.t.Error(msg)
	}
	crw.status = status
	crw.written = true
	crw.ResponseWriter.WriteHeader(status)
}

// RespWriterMW returns a [Middleware] that wraps the [http.ResponseWriter]
// in a [RespWriter] instance.
func RespWriterMW(t tester.T, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		next.ServeHTTP(NewRespWriter(t, w), r)
	})
}
