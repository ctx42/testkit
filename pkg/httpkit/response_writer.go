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

var _ http.ResponseWriter = (*RespWriter)(nil)

// NewRespWriter returns a new instance of [RespWriter].
func NewRespWriter(t tester.T, w http.ResponseWriter) *RespWriter {
	t.Helper()
	return &RespWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
		t:              t,
	}
}

// implements [http.ResponseWriter]. Errors the test if headers already sent.
func (rsw *RespWriter) Header() http.Header {
	if rsw.written {
		rsw.t.Helper()
		msg := notice.New("headers already written").
			Append("last status", "%d", rsw.status)
		rsw.t.Error(msg)
	}
	return rsw.ResponseWriter.Header()
}

func (rsw *RespWriter) Write(data []byte) (int, error) {
	rsw.written = true
	return rsw.ResponseWriter.Write(data)
}

// implements [http.ResponseWriter]. Errors the test if headers already sent.
func (rsw *RespWriter) WriteHeader(status int) {
	if rsw.written {
		rsw.t.Helper()
		msg := notice.New("headers already written").
			Append("last status", "%d", rsw.status).
			Append("new status", "%d", status)
		rsw.t.Error(msg)
	}
	rsw.status = status
	rsw.written = true
	rsw.ResponseWriter.WriteHeader(status)
}

// RespWriterMW returns a [Middleware] that wraps the [http.ResponseWriter]
// in a [RespWriter] instance.
func RespWriterMW(t tester.T, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		next.ServeHTTP(NewRespWriter(t, w), r)
	})
}
