package httpkit

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_WithRequestStatus(t *testing.T) {
	// --- Given ---
	opts := &Request{}

	// --- When ---
	WithRequestStatus(http.StatusBadRequest)(opts)

	// --- Then ---
	assert.Equal(t, http.StatusBadRequest, opts.expCode)
}

func Test_WithRequestTimeout(t *testing.T) {
	// --- Given ---
	opts := &Request{}

	// --- When ---
	WithRequestTimeout(time.Second)(opts)

	// --- Then ---
	assert.Equal(t, time.Second, opts.timeout)
}

func Test_WithRequestTries(t *testing.T) {
	// --- Given ---
	opts := &Request{}

	// --- When ---
	WithRequestTries(123)(opts)

	// --- Then ---
	assert.Equal(t, 123, opts.tries)
}

func Test_NewRequest(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		req := NewRequest(tspy)

		// --- Then ---
		assert.Equal(t, 3*time.Second, req.timeout)
		assert.Equal(t, http.StatusOK, req.expCode)
		assert.Same(t, tspy, req.t)
	})

	t.Run("custom timeout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		req := NewRequest(tspy, WithRequestTimeout(time.Second))

		// --- Then ---
		assert.Equal(t, time.Second, req.timeout)
	})
}

func Test_Request_Get(t *testing.T) {
	t.Run("success 200", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(req.URL.RawQuery))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).Get(srv.URL, "?key0=val0", "&key1=val1")

		// --- Then ---
		assert.Equal(t, "key0=val0&key1=val1", have)
	})

	t.Run("failure 400", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("FAIL"))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })

		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "expected HTTP response code:\n" +
			"\t url: %s\n" +
			"\tverb: GET\n" +
			"\twant: 200\n" +
			"\thave: 400"
		tspy.ExpectLogEqual(wMsg, srv.URL)
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).Get(srv.URL)

		// --- Then ---
		assert.Equal(t, "FAIL", have)
	})

	t.Run("invalid url", func(t *testing.T) {
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("invalid character \" \" in host name")
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).Get(" https://foo.com")

		// --- Then ---
		assert.Equal(t, "", have)
	})
}

func Test_Request_GetHeaders(t *testing.T) {
	t.Run("success 200", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(req.URL.RawQuery))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).GetHeaders(srv.URL, "?key0=val0", "&key1=val1")

		// --- Then ---
		delete(have, "Date")
		wHds := map[string][]string{
			"Content-Length": {"19"},
			"Content-Type":   {"text/plain; charset=utf-8"},
		}
		assert.Equal(t, wHds, have)
	})

	t.Run("invalid url", func(t *testing.T) {
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("invalid character \" \" in host name")
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).GetHeaders(" https://foo.com")

		// --- Then ---
		assert.Empty(t, have)
	})
}

func Test_Request_get(t *testing.T) {
	t.Run("success 200", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(req.URL.RawQuery))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).get(srv.URL, "?key0=val0", "&key1=val1")

		// --- Then ---
		hBody := must.Value(io.ReadAll(have.Body))
		assert.Equal(t, "key0=val0&key1=val1", string(hBody))
		assert.NoError(t, have.Body.Close())
	})

	t.Run("url without http protocol", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(req.URL.RawQuery))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })
		u := strings.Replace(srv.URL, "http://", "", 1)

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).get(u, "?key0=val0", "&key1=val1")

		// --- Then ---
		hBody := must.Value(io.ReadAll(have.Body))
		assert.Equal(t, "key0=val0&key1=val1", string(hBody))
		assert.NoError(t, have.Body.Close())
	})

	t.Run("success 400", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("OK"))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })

		tspy := tester.New(t)
		tspy.Close()

		req := NewRequest(tspy, WithRequestStatus(http.StatusBadRequest))

		// --- When ---
		have := req.get(srv.URL)

		// --- Then ---
		hBody := must.Value(io.ReadAll(have.Body))
		assert.Equal(t, "OK", string(hBody))
		assert.NoError(t, have.Body.Close())
	})

	t.Run("failure 400", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("FAIL"))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })

		tspy := tester.New(t)
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).get(srv.URL)

		// --- Then ---
		assert.Equal(t, "FAIL", string(must.Value(io.ReadAll(have.Body))))
		assert.NoError(t, have.Body.Close())
	})

	t.Run("timeout", func(t *testing.T) {
		// --- Given ---
		handler := func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusOK)
			time.Sleep(100 * time.Millisecond)
			_, _ = w.Write([]byte("OK"))
		}
		srv := httptest.NewServer(http.HandlerFunc(handler))
		t.Cleanup(func() { srv.Close() })

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("context deadline exceeded")
		tspy.Close()

		req := NewRequest(tspy, WithRequestTimeout(50*time.Millisecond))

		// --- When ---
		have := req.get(srv.URL) // nolint: bodyclose

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("connection refused", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(nil))
		url := srv.URL
		srv.Close()

		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogContain("connection refused")
		tspy.Close()

		req := NewRequest(tspy)

		// --- When ---
		have := req.get(url) // nolint: bodyclose

		// --- Then ---
		assert.Nil(t, have)
	})

	t.Run("invalid url", func(t *testing.T) {
		tspy := tester.New(t)
		tspy.ExpectError()
		tspy.ExpectLogEqual("Get \"https:\\\\\": http: no Host in request URL")
		tspy.Close()

		// --- When ---
		have := NewRequest(tspy).get("https:\\") // nolint: bodyclose

		// --- Then ---
		assert.Nil(t, have)
	})
}

func Test_IsConnRefused(t *testing.T) {
	t.Run("is", func(t *testing.T) {
		// --- Given ---
		err := &net.OpError{
			Err: &os.SyscallError{
				Err: syscall.ECONNREFUSED,
			},
		}

		// --- When ---
		have := IsConnRefused(err)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("is not", func(t *testing.T) {
		// --- When ---
		have := IsConnRefused(errors.New("test error"))

		// --- Then ---
		assert.False(t, have)
	})

	t.Run("nil", func(t *testing.T) {
		// --- When ---
		have := IsConnRefused(nil)

		// --- Then ---
		assert.False(t, have)
	})
}
