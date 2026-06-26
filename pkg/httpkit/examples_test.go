// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package httpkit_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/ctx42/testkit/pkg/httpkit"
)

func ExampleHandleFunc() {
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
}

func ExampleNewServer() {
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
}

func ExampleServer_BodyString() {
	t := &testing.T{}
	srv := httpkit.NewServer(t).Rsp(http.StatusOK, nil)

	r, _ := http.Post(srv.URL(), "", strings.NewReader("ping"))
	defer func() { _ = r.Body.Close() }()
	fmt.Println(srv.BodyString(0))
	// Output:
	// ping
}

func ExampleRequest_Get() {
	t := &testing.T{}
	srv := httpkit.NewServer(t).Rsp(http.StatusOK, []byte("pong"))

	body := httpkit.NewRequest(t).Get(srv.URL())
	fmt.Println(body)
	// Output:
	// pong
}
