// Package httpkit provides test helpers for HTTP handler and server
// testing. It offers three main tools:
//
//   - [Handler] — a thin wrapper around [httptest.Server] for testing
//     specific HTTP handler functions with optional middleware.
//   - [Server] — a request-recording test server that returns
//     pre-configured responses in order.
//   - [Request] — an outbound HTTP client for exercising real servers
//     from within tests.
package httpkit

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/ctx42/testing/pkg/tester"
)

// Middleware wraps an [http.Handler] to extend its behaviour.
type Middleware func(next http.Handler) http.Handler

// Noop is no operation handler function that writes HTTP 204 response.
func Noop(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// NoopHandler is no operation handler.
var NoopHandler = http.HandlerFunc(Noop)

// Handler is a thin wrapper around [httptest.Server].
type Handler struct {
	*httptest.Server          // Wrapped test server.
	t                tester.T // Test manager.
}

// HandleFunc returns new [Handler] serving given handler function at the given
// pattern. You need to call [Handler.Start] or [Handler.StartTLS] to actually
// start it. The pattern defaults to "/{$}" when the provided pattern is empty
// or "/". The started server is automatically closed at the test end.
func HandleFunc(
	t tester.T,
	pattern string,
	fn http.HandlerFunc,
	mws ...Middleware,
) *Handler {

	var h http.Handler
	h = RespWriterMW(t, fn)
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	mux := http.NewServeMux()
	if pattern == "/" || pattern == "" {
		pattern = "/{$}"
	}
	mux.Handle(pattern, h)
	han := &Handler{Server: httptest.NewUnstartedServer(mux), t: t}
	t.Cleanup(func() { t.Helper(); han.Close() })
	return han
}

// Handle returns new [Handler] serving given handler. You need to call
// [Handler.Start] or [Handler.StartTLS] to actually start it. The started
// server is automatically closed at the test end.
func Handle(t tester.T, hun http.Handler) *Handler {
	han := &Handler{Server: httptest.NewUnstartedServer(hun), t: t}
	t.Cleanup(func() { t.Helper(); han.Close() })
	return han
}

// Start calls [httptest.Server.Start] and provides a fluent interface. The
// provided context may be set to nil if it's unnecessary.
func (han *Handler) Start(ctx context.Context) *Handler {
	han.t.Helper()
	if ctx != nil {
		han.Config.BaseContext = func(_ net.Listener) context.Context {
			return ctx
		}
	}
	han.Server.Start()
	return han
}

// StartTLS calls [httptest.Server.StartTLS] and provides a fluent interface.
// The provided context may be set to nil if it's unnecessary.
func (han *Handler) StartTLS(ctx context.Context) *Handler {
	han.t.Helper()
	if ctx != nil {
		han.Config.BaseContext = func(_ net.Listener) context.Context {
			return ctx
		}
	}
	han.Server.StartTLS()
	return han
}

// Info returns server connection information.
func (han *Handler) Info() map[string]any {
	han.t.Helper()
	u, err := url.Parse(han.URL)
	if err != nil {
		han.t.Error(err)
		return nil
	}
	return map[string]any{
		"scheme": u.Scheme, // Test server HTTP scheme.
		"host":   u.Host,   // Test server host and port.
		"port":   u.Port(), // Test server port.
		"url":    han.URL,  // Test server URL.
	}
}
