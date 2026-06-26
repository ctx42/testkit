package httpkit

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"

	"github.com/ctx42/testkit/pkg/iokit"
)

func Test_Server_smokeTest(t *testing.T) {
	t.Run("smoke test", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		srv := NewServer(tspy).Rsp(http.StatusOK, []byte("response"))

		// --- When ---
		body := bytes.NewReader([]byte("req body"))
		rsp, err := http.Post(srv.URL()+"/?k0=v0", "", body) // nolint:noctx

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rsp.StatusCode)
		assert.Equal(t, "response", iokit.ReadAllStr(t, rsp.Body))
		assert.Equal(t, 1, srv.ReqCount())
		assert.Equal(t, "req body", string(srv.Body(0)))
		assert.Len(t, 1, srv.Values(0))
		assert.Equal(t, "v0", srv.Values(0).Get("k0"))
		assert.NoError(t, rsp.Body.Close())

		req := srv.Request(0)

		u := must.Value(url.Parse(srv.URL()))
		assert.Equal(t, u.String()+"/?k0=v0", req.URL.String())
		assert.Equal(t, "req body", iokit.ReadAllStr(t, req.Body))
		_, port, err := net.SplitHostPort(u.Host)
		assert.NoError(t, err)
		assert.Equal(t, port, srv.Port())
		assert.Equal(t, u.Host, srv.Host())
	})

	t.Run("response with delay", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		srv := NewServer(tspy).
			Rsp(http.StatusOK, []byte("response")).Delay(time.Second)

		// --- When ---
		start := time.Now()
		body := bytes.NewReader([]byte("req body"))
		rsp, err := http.Post(srv.URL(), "", body) // nolint:noctx
		took := time.Since(start)

		// --- Then ---
		if assert.NoError(t, err) {
			defer func() { _ = rsp.Body.Close() }()
		}
		assert.True(t, took > time.Second)
	})

	t.Run("response with custom header", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		srv := NewServer(tspy).
			Rsp(http.StatusOK, []byte("response")).
			Header("X-Mine", "ABC")

		// --- When ---
		body := bytes.NewReader([]byte("req body"))
		rsp, err := http.Post(srv.URL(), "", body) // nolint:noctx

		// --- Then ---
		if assert.NoError(t, err) {
			defer func() { _ = rsp.Body.Close() }()
		}
		assert.Equal(t, "ABC", rsp.Header.Get("X-Mine"))
	})

	t.Run("empty body", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		srv := NewServer(tspy).Rsp(http.StatusCreated, nil)

		// --- When ---
		body := bytes.NewReader([]byte("req body"))
		rsp, err := http.Post(srv.URL()+"/?k0=v0", "", body) // nolint:noctx

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rsp.StatusCode)
		assert.Equal(t, "", iokit.ReadAllStr(t, rsp.Body))
		assert.Equal(t, 1, srv.ReqCount())
		assert.NoError(t, rsp.Body.Close())

		req := srv.Request(0)
		assert.Equal(t, srv.URL()+"/?k0=v0", req.URL.String())
		assert.Equal(t, "req body", iokit.ReadAllStr(t, req.Body))
	})

	t.Run("number of responses do not match requests", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFail()
		tspy.ExpectCleanups(1)
		tspy.ExpectLogEqual("expected %d requests got %d", 2, 1)
		tspy.Close()

		// Expect two requests.
		srv := NewServer(tspy)
		srv.Rsp(http.StatusOK, nil)
		srv.Rsp(http.StatusOK, nil)

		// --- When ---
		rsp, err := http.Get(srv.URL()) // nolint:noctx

		// --- Then ---
		if assert.NoError(t, err) {
			defer func() { _ = rsp.Body.Close() }()
		}
		// This triggers Errorf in cleanup function.
		tspy.Finish().AssertExpectations()
	})

	t.Run("request timeout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		// Expect two requests.
		srv := NewServer(tspy)
		srv.Rsp(http.StatusOK, nil).Delay(time.Second)

		// --- When ---
		ctx := context.Background()
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			srv.URL(),
			http.NoBody,
		)
		assert.NoError(t, err)

		client := &http.Client{Timeout: 500 * time.Millisecond}
		rsp, err := client.Do(req)
		if !assert.Nil(t, rsp) {
			assert.NoError(t, rsp.Body.Close())
		}
		var e net.Error
		assert.ErrorAs(t, &e, err)
		assert.True(t, e.Timeout())

		// --- Then ---
		tspy.Finish().AssertExpectations()
	})
}
