package httpkit

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/ctx42/testing/pkg/tester"
)

// response represents an HTTP response returned by Server.
type response struct {
	status  int
	body    []byte
	headers http.Header
	delay   time.Duration
}

// Server represents a very basic HTTP server recording all the requests it
// receives, responding with responses added with [Server.Rsp] method.
//
// The server instance provides methods to analyze the received requests.
type Server struct {
	srv         *httptest.Server // The test server.
	host        string           // Test server host:port.
	port        string           // Test server port.
	scheme      string           // Test server scheme.
	requestCnt  int              // Number of received requests.
	responseIdx int              // Index of last-returned response from responses.
	responses   []response       // Responses to return.
	requests    []*http.Request  // Received requests.
	t           tester.T         // Test state manager.
}

// NewServer returns a new instance of [Server], and registers a call to Close
// in test cleanup. The server will fail the test if during cleanup the number
// of expected responses does not match the number of seen requests.
func NewServer(t tester.T) *Server {
	t.Helper()
	srv := &Server{
		responseIdx: -1,
		t:           t,
	}

	// Cleanup after the test is done.
	t.Cleanup(func() {
		t.Helper()
		if srv.requestCnt != len(srv.responses) {
			t.Errorf(
				"expected %d requests got %d",
				len(srv.responses),
				srv.requestCnt,
			)
		}
		_ = srv.Close()
	})

	// Handler for all incoming requests.
	handler := func(w http.ResponseWriter, req *http.Request) {
		srv.requestCnt++
		rsp := srv.next()
		if rsp.delay != 0 {
			time.Sleep(rsp.delay)
		}

		for key, vs := range rsp.headers {
			for _, v := range vs {
				w.Header().Add(key, v)
			}
		}
		w.WriteHeader(rsp.status)

		c := cloneHTTPRequest(t, req)
		c.URL.Scheme = srv.scheme

		srv.requests = append(srv.requests, c)
		if rsp.body != nil {
			if _, err := w.Write(rsp.body); err != nil {
				t.Error(err)
				return
			}
		}
	}

	srv.srv = httptest.NewServer(http.HandlerFunc(handler))
	u, err := url.Parse(srv.srv.URL)
	if err != nil {
		t.Error(err)
		return nil
	}
	srv.host = u.Host
	srv.scheme = u.Scheme
	srv.port = u.Port()

	return srv
}

// Rsp adds a response to the queue. Every time the server receives a request
// it returns the next queued response in the order they were added. Pass nil
// for rsp to send a response with no body.
func (srv *Server) Rsp(status int, rsp []byte) *Server {
	srv.responses = append(srv.responses, response{
		status: status,
		body:   rsp,
	})
	return srv
}

// Delay configures delay for the last defined response. On error, it marks the
// test as failed and returns nil.
func (srv *Server) Delay(delay time.Duration) *Server {
	if len(srv.responses) == 0 {
		srv.t.Error("you need to define response first")
		return nil
	}
	srv.responses[len(srv.responses)-1].delay = delay
	return srv
}

// Header adds header for the last defined response. On error, it marks the
// test as failed and returns nil.
func (srv *Server) Header(key, value string) *Server {
	if len(srv.responses) == 0 {
		srv.t.Error("you need to define a response first")
		return nil
	}
	rsp := srv.responses[len(srv.responses)-1]
	if rsp.headers == nil {
		srv.responses[len(srv.responses)-1].headers = make(http.Header)
	}
	srv.responses[len(srv.responses)-1].headers.Add(key, value)
	return srv
}

// URL returns URL for the test server.
func (srv *Server) URL() string { return srv.srv.URL }

// Port returns the port of the test server.
func (srv *Server) Port() string { return srv.port }

// Host returns the host address of the test server (host:port).
func (srv *Server) Host() string { return srv.host }

// Request returns the clone of the nth received request. If n is greater than
// the number of received requests, it marks the test as failed and returns nil.
func (srv *Server) Request(n int) *http.Request {
	srv.t.Helper()
	if n >= 0 && n < len(srv.requests) {
		return cloneHTTPRequest(srv.t, srv.requests[n])
	}
	srv.t.Errorf("no request with index %d recorded", n)
	return nil
}

// ReqCount returns the number of requests recorded by the test server.
func (srv *Server) ReqCount() int { return len(srv.requests) }

// Values returns URL query values of the nth received request. On error, it
// marks the test as failed and returns nil.
func (srv *Server) Values(n int) url.Values {
	srv.t.Helper()
	if n >= 0 && n < len(srv.requests) {
		return srv.requests[n].URL.Query()
	}
	srv.t.Errorf("no request with index %d recorded", n)
	return nil
}

// Body returns body of the nth received request. On error, it marks the test
// as failed and returns nil.
func (srv *Server) Body(n int) []byte {
	srv.t.Helper()
	if n >= 0 && n < len(srv.requests) {
		req := srv.requests[n]
		var buf bytes.Buffer
		body, err := io.ReadAll(io.TeeReader(req.Body, &buf))
		if err != nil {
			srv.t.Error(err)
			return nil
		}
		if err = req.Body.Close(); err != nil {
			srv.t.Error(err)
			return nil
		}
		req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		return body
	}
	srv.t.Errorf("no request with index %d recorded", n)
	return nil
}

// BodyString returns the body of the nth received request. On error, it marks
// the test as failed and returns an empty string.
func (srv *Server) BodyString(n int) string {
	srv.t.Helper()
	return string(srv.Body(n))
}

// Headers returns headers for given request index. On error, it marks the test
// as failed and returns nil.
func (srv *Server) Headers(n int) http.Header {
	srv.t.Helper()
	if n >= 0 && n < len(srv.requests) {
		return srv.requests[n].Header
	}
	srv.t.Errorf("no request with index %d recorded", n)
	return nil
}

// next returns the next response to return. On error, it marks the test as
// failed and returns an empty response struct.
func (srv *Server) next() response {
	srv.t.Helper()
	srv.responseIdx++ // Next response to give.
	var rsp response
	if srv.responseIdx >= len(srv.responses) {
		srv.t.Error("no more responses to give")
		return rsp
	}
	return srv.responses[srv.responseIdx]
}

// Close stops the test server and does cleanup. May be called multiple times.
// Never returns error.
func (srv *Server) Close() error {
	srv.srv.Close()
	for _, req := range srv.requests {
		_ = req.Body.Close()
	}
	srv.requests = srv.requests[:0]
	srv.responses = srv.responses[:0]
	return nil
}

// cloneHTTPRequest clones HTTP request and body, URL.Host and URL.Scheme.
// On error, it marks the test as failed and returns nil.
func cloneHTTPRequest(t tester.T, req *http.Request) *http.Request {
	t.Helper()
	c := req.Clone(context.Background())
	if req.Body != nil {
		var buf bytes.Buffer
		body, err := io.ReadAll(io.TeeReader(req.Body, &buf))
		if err != nil {
			t.Error(err)
			return nil
		}
		if err := req.Body.Close(); err != nil {
			t.Error(err)
			return nil
		}
		req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))
		c.Body = io.NopCloser(bytes.NewReader(body))
	}
	c.URL.Host = req.Host
	c.URL.Scheme = req.URL.Scheme
	return c
}
