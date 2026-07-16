// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package httpkit

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
)

// cKey is used as a typed context value key.
type cKey string

// CALL represents context value key.
const CALL cKey = "CALL"

func Test_Noop(t *testing.T) {
	// --- Given ---
	rec := httptest.NewRecorder()

	// --- When ---
	Noop(rec, nil)

	// --- Then ---
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "", rec.Body.String())
}

func Test_NoopHandler(t *testing.T) {
	// --- Given ---
	rec := httptest.NewRecorder()

	// --- When ---
	NoopHandler.ServeHTTP(rec, nil)

	// --- Then ---
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "", rec.Body.String())
}

func Test_HandleFunc(t *testing.T) {
	t.Run("success request", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value("CALL")
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", val))
		}
		han := HandleFunc(tspy, "/", fn)

		// --- When ---
		han.Start(nil)

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		assert.Equal(t, []string{"fn", "<nil>"}, rsp.Header.Values("CALL"))
	})

	t.Run("success TLS request", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value("CALL")
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", val))
		}
		han := HandleFunc(tspy, "/", fn)

		// --- When ---
		han.StartTLS(nil)

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		assert.Equal(t, []string{"fn", "<nil>"}, rsp.Header.Values("CALL"))
	})

	t.Run("success request with middleware", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
		}
		mw0 := func(next http.Handler) http.Handler {
			hf := func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				w.Header().Add("CALL", "mw0")
				w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(hf)
		}
		mw1 := func(next http.Handler) http.Handler {
			hf := func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				w.Header().Add("CALL", "mw1")
				w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(hf)
		}
		han := HandleFunc(tspy, "/", fn, mw0, mw1)

		// --- When ---
		han.Start(nil)

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		want := []string{
			"mw0", "<nil>",
			"mw1", "<nil>",
			"fn", "<nil>",
		}
		assert.Equal(t, want, rsp.Header.Values("CALL"))
	})

	t.Run("success context is propagated", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
		}
		mw0 := func(next http.Handler) http.Handler {
			hf := func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				w.Header().Add("CALL", "mw0")
				w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(hf)
		}
		han := HandleFunc(tspy, "/", fn, mw0)

		// --- When ---
		han.Start(context.WithValue(context.Background(), CALL, "val"))

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		want := []string{
			"mw0", "val",
			"fn", "val",
		}
		assert.Equal(t, want, rsp.Header.Values("CALL"))
	})

	t.Run("success context is propagated when TLS used", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
		}
		mw0 := func(next http.Handler) http.Handler {
			hf := func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				w.Header().Add("CALL", "mw0")
				w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
				next.ServeHTTP(w, r)
			}
			return http.HandlerFunc(hf)
		}
		han := HandleFunc(tspy, "/", fn, mw0)

		// --- When ---
		han.StartTLS(context.WithValue(context.Background(), CALL, "val"))

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		want := []string{
			"mw0", "val",
			"fn", "val",
		}
		assert.Equal(t, want, rsp.Header.Values("CALL"))
	})

	t.Run("server closed at test end", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value("CALL")
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", val))
		}
		han := HandleFunc(tspy, "/", fn)
		han.Start(nil)

		// --- When ---
		tspy.Finish()

		// --- Then ---
		_, err := han.Client().Get(han.URL)
		assert.ErrorContain(t, "connection refused", err)
	})
}

func Test_Handle(t *testing.T) {
	t.Run("success request", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value("CALL")
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", val))
		}
		han := Handle(tspy, http.HandlerFunc(fn))

		// --- When ---
		han.Start(nil)

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		assert.Equal(t, []string{"fn", "<nil>"}, rsp.Header.Values("CALL"))
	})

	t.Run("success TLS request", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value("CALL")
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", val))
		}
		han := Handle(tspy, http.HandlerFunc(fn))

		// --- When ---
		han.StartTLS(nil)

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		assert.Equal(t, []string{"fn", "<nil>"}, rsp.Header.Values("CALL"))
	})

	t.Run("success context is propagated", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
		}
		han := Handle(tspy, http.HandlerFunc(fn))

		// --- When ---
		han.Start(context.WithValue(context.Background(), CALL, "val"))

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		assert.Equal(t, []string{"fn", "val"}, rsp.Header.Values("CALL"))
	})

	t.Run("success context is propagated when TLS used", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", ctx.Value(CALL)))
		}
		han := Handle(tspy, http.HandlerFunc(fn))

		// --- When ---
		han.StartTLS(context.WithValue(context.Background(), CALL, "val"))

		// --- Then ---
		rsp, err := han.Client().Get(han.URL)
		assert.NoError(t, err)
		assert.NoError(t, rsp.Body.Close())
		assert.Equal(t, []string{"fn", "val"}, rsp.Header.Values("CALL"))
	})

	t.Run("server closed at test end", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value("CALL")
			w.Header().Add("CALL", "fn")
			w.Header().Add("CALL", fmt.Sprintf("%v", val))
		}
		han := Handle(tspy, http.HandlerFunc(fn))
		han.Start(nil)

		// --- When ---
		tspy.Finish()

		// --- Then ---
		_, err := han.Client().Get(han.URL)
		assert.ErrorContain(t, "connection refused", err)
	})
}

func Test_Handler_Info(t *testing.T) {
	t.Run("no https", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("CALL", "fn")
		}
		han := HandleFunc(tspy, "/", fn).Start(nil)

		// --- When ---
		have := han.Info()

		// --- Then ---
		assert.Equal(t, "http", have["scheme"])
		assert.HasKey(t, "host", have)
		assert.HasKey(t, "port", have)
		assert.HasKey(t, "url", have)
	})

	t.Run("with https", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("CALL", "fn")
		}
		han := HandleFunc(tspy, "/", fn).StartTLS(nil)

		// --- When ---
		have := han.Info()

		// --- Then ---
		assert.Equal(t, "https", have["scheme"])
		assert.HasKey(t, "host", have)
		assert.HasKey(t, "port", have)
		assert.HasKey(t, "url", have)
	})
}
