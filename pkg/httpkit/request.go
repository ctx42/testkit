// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package httpkit

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/ctx42/testing/pkg/tester"
)

// WithRequestStatus is a [NewRequest] option setting the expected HTTP
// response status code.
func WithRequestStatus(code int) func(*Request) {
	return func(req *Request) { req.expCode = code }
}

// WithRequestTimeout is a [NewRequest] option setting the request timeout.
func WithRequestTimeout(to time.Duration) func(*Request) {
	return func(req *Request) { req.timeout = to }
}

// WithRequestTries is a [NewRequest] option setting the number of retries when
// request experiences "connection refused" error.
func WithRequestTries(n int) func(opt *Request) {
	return func(req *Request) { req.tries = n }
}

// Request represents HTTP request.
type Request struct {
	timeout time.Duration // Request timeout.
	tries   int           // Request tries on "connection refused" (default: 1).
	tryWait time.Duration // The "connection refused" wait (default: 10ms).
	expCode int           // Expected HTTP status code.
	t       tester.T      // Test manager.
}

// NewRequest returns a new instance of Request.
func NewRequest(t tester.T, opts ...func(*Request)) *Request {
	t.Helper()
	tst := &Request{
		timeout: 3 * time.Second,
		expCode: 200,
		tries:   1,
		tryWait: 10 * time.Millisecond,
		t:       t,
	}
	for _, opt := range opts {
		opt(tst)
	}
	return tst
}

// Get makes HTTP GET request to given url and returns the body. On error, it
// marks the test as failed and returns an empty string.
func (tst *Request) Get(u string, parts ...string) string {
	tst.t.Helper()
	rsp := tst.get(u, parts...)
	if tst.t.Failed() {
		return ""
	}
	defer func() { _ = rsp.Body.Close() }()

	data, err := io.ReadAll(rsp.Body)
	if err != nil {
		tst.t.Error(err)
		return ""
	}

	if tst.expCode != rsp.StatusCode {
		const format = "expected HTTP response code:\n" +
			"\t url: %s\n" +
			"\tverb: GET\n" +
			"\twant: %d\n" +
			"\thave: %d"
		tst.t.Errorf(format, u, tst.expCode, rsp.StatusCode)
		return string(data)
	}
	return string(data)
}

// GetHeaders makes HTTP GET request to given url and returns headers
// discarding the body. On error, it marks the test as failed and returns nil.
func (tst *Request) GetHeaders(u string, parts ...string) map[string][]string {
	tst.t.Helper()
	rsp := tst.get(u, parts...)
	if tst.t.Failed() {
		return nil
	}
	defer func() { _ = rsp.Body.Close() }()
	if _, err := io.Copy(io.Discard, rsp.Body); err != nil {
		tst.t.Error(err)
		return nil
	}
	return rsp.Header
}

// get makes HTTP GET request to given url and returns the response. On error,
// it marks the test as failed and returns an empty string.
func (tst *Request) get(u string, parts ...string) *http.Response {
	tst.t.Helper()
	ctx := context.Background()

	var cxl func()
	if tst.timeout > 0 {
		ctx, cxl = context.WithTimeout(ctx, tst.timeout)
		defer cxl()
	}

	if !strings.HasPrefix(u, "http") {
		u = "http://" + u
	}

	u += strings.Join(parts, "")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		tst.t.Error(err)
		return nil
	}

	tries := tst.tries
	for {
		var rsp *http.Response
		if rsp, err = http.DefaultClient.Do(req); err == nil {
			return rsp
		}
		if tries > 0 && IsConnRefused(err) {
			time.Sleep(tst.tryWait)
			tries--
			continue
		}
		tst.t.Error(err)
		return nil
	}
}

// IsConnRefused returns true if the error is a connection-refused error.
func IsConnRefused(err error) bool {
	if err == nil {
		return false
	}
	if opErr, ok := errors.AsType[*net.OpError](err); ok {
		if sysErr, ok2 := errors.AsType[*os.SyscallError](opErr.Err); ok2 {
			return errors.Is(sysErr.Err, syscall.ECONNREFUSED)
		}
	}
	return false
}
