// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package netkit provides networking helpers for tests: obtaining
// free TCP ports and local addresses, checking whether a TCP address
// accepts connections, and skipping tests when no network or Internet
// connection is available.
//
// Most helpers return their result (and an error where one is
// possible) for the caller to handle. CanConnect, CanNotConnect, and
// SkipOnNoNetConn take a [tester.T] and report failures through it.
package netkit

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ctx42/testing/pkg/notice"
	"github.com/ctx42/testing/pkg/tester"
)

// GetFreePort asks the kernel for a free open port that is ready to use.
//
// The solution is to bind to port 0, which asks the kernel to allocate a port
// from /proc/sys/net/ipv4/ip_local_port_range. Then, close the socket
// and use that port number.
//
// This works because the kernel doesn't seem to reuse port numbers until it
// absolutely has to. The subsequent binds to port 0 will allocate a different
// port number.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil // nolint:forcetypeassert
}

// GetFreePorts calls [GetFreePort] count times and returns the
// allocated ports.
func GetFreePorts(count int) ([]int, error) {
	var ports []int
	for range count {
		port, err := GetFreePort()
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}
	return ports, nil
}

// GetLocalAddress works like [GetFreePort] but returns port with localhost.
// You may provide a prefix (the first value).
func GetLocalAddress(prefix ...string) (string, error) {
	var pref string
	if len(prefix) > 0 {
		pref = prefix[0]
	}
	port, err := GetFreePort()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%slocalhost:%d", pref, port), nil
}

// ReservePort reserves host port.
func ReservePort(port int) error {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	_ = l.Close()
	return nil
}

// GetFreePortRange reserves n consecutive host ports with the maximum number
// of retries. It is strongly advised to use n < 10 and retry < 5.
func GetFreePortRange(n, retry int) ([]int, error) {
	if retry == 0 {
		return nil, errors.New("could not reserve port range")
	}
	port, err := GetFreePort()
	if err != nil {
		return nil, err
	}

	reserved := make([]int, 0, n)
	reserved = append(reserved, port)
	for i := port + 1; i < port+n; i++ {
		if err = ReservePort(i); err != nil {
			return GetFreePortRange(n, retry-1)
		}
		reserved = append(reserved, i)
	}
	return reserved, nil
}

// RandomPorts returns n consecutive ports with random beginning.
func RandomPorts(n int) []string {
	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	start := rnd.Intn(65535-50000-n) + 50000
	var ret []string
	for i := start; i < start+n; i++ {
		ret = append(ret, strconv.Itoa(i))
	}
	return ret
}

// CanConnect attempts to establish a TCP connection to the address with the
// given timeout. On success, it returns the established connection, otherwise
// marks the test as failed, writes an error message to the test log and
// returns nil.
func CanConnect(t tester.T, timeout, address string) net.Conn {
	t.Helper()
	to, err := time.ParseDuration(timeout)
	if err != nil {
		msg := notice.New("expected valid time duration").
			Have("%s", timeout).
			Append("error", "%s", err)
		t.Fatal(msg)
		return nil
	}
	conn, err := net.DialTimeout("tcp", address, to)
	if err != nil {
		msg := notice.New("expected successful connection").
			Append("error", "%s", err)
		t.Fatal(msg)
		return nil
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// CanNotConnect attempts to establish a TCP connection to the address. On
// "connection refused" it returns true, otherwise marks the test as failed,
// writes an error message to the test log and returns false.
func CanNotConnect(t tester.T, address string) bool {
	t.Helper()
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		var e *net.OpError
		if errors.As(err, &e) {
			have := e.Err.Error()
			if e.Op == "dial" && have == "connect: connection refused" {
				return true
			}
		}
		msg := notice.New("expected connection refused error").
			Have("%s", err).
			Append("address", "%s", address)
		t.Error(msg)
		return false
	}
	_ = conn.Close()
	msg := notice.New("expected no connection possible").
		Append("address", "%s", address)
	t.Error(msg)
	return false
}

// GetLocalIP returns local IP which is most likely used to connect to the
// Internet. It creates a UDP connection (doesn't send any data) to a Google
// Public DNS server and then retrieves the local address. The IP of the local
// address is the preferred outbound IP address of your machine. On error, it
// returns loopback (127.0.0.1) address.
func GetLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.IPv4(127, 0, 0, 1)
	}
	defer func() { _ = conn.Close() }()
	addr := conn.LocalAddr().(*net.UDPAddr) // nolint: forcetypeassert
	return addr.IP
}

// Makes sure Internet connection is checked only once.
var chk sync.Once

// isConnected is true if the Internet connection is available.
var isConnected bool

// HasNetConn returns true if the Internet connection is present.
func HasNetConn() bool {
	chk.Do(func() {
		if _, err := net.LookupIP("www.google.com"); err == nil {
			isConnected = true
		}
	})
	return isConnected
}

// SkipOnNoNetConn skips the test if there is no Internet connection.
func SkipOnNoNetConn(t tester.T) {
	t.Helper()
	if !HasNetConn() {
		t.Skip("skipping test: no Internet connection")
	}
}
