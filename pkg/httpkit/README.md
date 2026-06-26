<!-- TOC -->
* [The `httpkit` package](#the-httpkit-package)
  * [Handler](#handler)
    * [HandleFunc](#handlefunc)
    * [Handle](#handle)
  * [Server](#server)
    * [NewServer / Rsp](#newserver--rsp)
    * [Delay / Header](#delay--header)
    * [Inspecting recorded requests](#inspecting-recorded-requests)
  * [Request](#request)
    * [Get / GetHeaders](#get--getheaders)
<!-- TOC -->

# The `httpkit` package

`httpkit` provides test helpers for HTTP handler and server testing.
Every helper integrates with `tester.T`: on error it marks the test as
failed, writes a diagnostic to the test log, and returns a safe zero
value so the calling test can continue executing.

## Handler

`Handler` wraps [httptest.Server] and is created by `HandleFunc` or
`Handle`. The server is not started until `Start` or `StartTLS` is
called, giving you time to configure middleware. It is automatically
closed at the end of the test.

### HandleFunc

Registers a single `http.HandlerFunc` at the given pattern, wraps it
in `RespWriterMW`, and returns an unstarted `Handler`. Call `Start`
(or `StartTLS`) to make it reachable.

<!-- gmdoceg:ExampleHandleFunc -->
```go
t := &testing.T{}
fn := func(w http.ResponseWriter, _ *http.Request) {
    _, _ = fmt.Fprintln(w, "hello")
}
han := httpkit.HandleFunc(t, "/", fn).Start(nil)

rsp, _ := http.Get(han.URL + "/")
defer func() { _ = rsp.Body.Close() }()
body, _ := io.ReadAll(rsp.Body)
fmt.Print(string(body))
// Output:
// hello
```

### Handle

`Handle` accepts a fully assembled `http.Handler` — useful when you
need a custom mux or multiple routes.

```go
mux := http.NewServeMux()
mux.HandleFunc("/a", handlerA)
mux.HandleFunc("/b", handlerB)

han := httpkit.Handle(t, mux).Start(nil)
```

---

## Server

`Server` is a request-recording HTTP server. You queue responses with
`Rsp` before the test runs; the server returns them in order as
requests arrive. At test cleanup it fails the test if the number of
queued responses does not match the number of received requests.

### NewServer / Rsp

<!-- gmdoceg:ExampleNewServer -->
```go
t := &testing.T{}
srv := httpkit.NewServer(t).Rsp(http.StatusOK, []byte("pong"))

rsp, _ := http.Get(srv.URL())
defer func() { _ = rsp.Body.Close() }()
body, _ := io.ReadAll(rsp.Body)
fmt.Println(rsp.StatusCode)
fmt.Println(string(body))
// Output:
// 200
// pong
```

### Delay / Header

Fluent modifiers that apply to the most recently added response:

```go
srv := httpkit.NewServer(t).
    Rsp(http.StatusOK, []byte("slow")).Delay(200 * time.Millisecond).
    Rsp(http.StatusCreated, nil).Header("X-ID", "42")
```

### Inspecting recorded requests

Each received request is stored as a clone. Use `Request(n)`, `Body(n)`
/ `BodyString(n)`, `Headers(n)`, and `Values(n)` to inspect them.

<!-- gmdoceg:ExampleServer_BodyString -->
```go
t := &testing.T{}
srv := httpkit.NewServer(t).Rsp(http.StatusOK, nil)

r, _ := http.Post(srv.URL(), "", strings.NewReader("ping"))
defer func() { _ = r.Body.Close() }()
fmt.Println(srv.BodyString(0))
// Output:
// ping
```

---

## Request

`Request` is a thin outbound HTTP client for exercising real servers
from within tests. It retries on connection-refused errors, enforces a
timeout, and fails the test if the actual response status does not
match the expected one.

### Get / GetHeaders

`Get` returns the response body as a string. `GetHeaders` discards the
body and returns the response headers.

<!-- gmdoceg:ExampleRequest_Get -->
```go
t := &testing.T{}
srv := httpkit.NewServer(t).Rsp(http.StatusOK, []byte("pong"))

body := httpkit.NewRequest(t).Get(srv.URL())
fmt.Println(body)
// Output:
// pong
```

Use `WithRequestStatus`, `WithRequestTimeout`, and `WithRequestTries`
to override the defaults:

```go
req := httpkit.NewRequest(t,
    httpkit.WithRequestStatus(http.StatusCreated),
    httpkit.WithRequestTimeout(5*time.Second),
    httpkit.WithRequestTries(3),
)
```
