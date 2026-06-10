// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package netkit

import (
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testing/pkg/tester"
)

func Test_GetFreePorts(t *testing.T) {
	// --- When ---
	ports, err := GetFreePorts(3)

	// --- Then ---
	assert.NoError(t, err)
	check := make(map[int]struct{}, 3)
	for i := range ports {
		if _, ok := check[i]; ok {
			t.Error("duplicate ports on the list")
		}
		check[i] = struct{}{}
	}
}

func Test_GetFreePortRange(t *testing.T) {
	// --- When ---
	n := 3
	ports, err := GetFreePortRange(n, 3)

	// --- Then ---
	assert.NoError(t, err)
	for i := 0; i < n-1; i++ {
		if ports[i]+1 != ports[i+1] {
			t.Error("ports are not consecutive numbers")
		}
	}
}

func Test_RandomPorts(t *testing.T) {
	// --- When ---
	ports := RandomPorts(3)

	// --- Then ---
	assert.Len(t, 3, ports)
	start, err := strconv.Atoi(ports[0])
	assert.NoError(t, err)
	assert.True(t, start > 50000)
	assert.Equal(t, strconv.Itoa(start+1), ports[1])
	assert.Equal(t, strconv.Itoa(start+2), ports[2])
}

func Test_GetLocalAddress(t *testing.T) {
	t.Run("without prefix", func(t *testing.T) {
		// --- When ---
		have, err := GetLocalAddress()

		// --- Then ---
		assert.NoError(t, err)
		host, port := must.Values(net.SplitHostPort(have))
		assert.Equal(t, "localhost", host)
		assert.NotEmpty(t, port)
	})

	t.Run("with prefix", func(t *testing.T) {
		// --- When ---
		have, err := GetLocalAddress("http://", "abc")

		// --- Then ---
		assert.NoError(t, err)
		port, ok := strings.CutPrefix(have, "http://localhost:")
		assert.True(t, ok)
		assert.NotEmpty(t, port)
	})
}

func Test_CanConnect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		lnr := must.Value(net.Listen("tcp", "127.0.0.1:0"))
		t.Cleanup(func() { _ = lnr.Close() })

		// --- When ---
		have := CanConnect(tspy, "1s", lnr.Addr().String())

		// --- Then ---
		assert.NotNil(t, have)
		assert.NoError(t, have.Close())
	})

	t.Run("Close called at test end", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectCleanups(1)
		tspy.Close()

		lnr := must.Value(net.Listen("tcp", "127.0.0.1:0"))
		t.Cleanup(func() { _ = lnr.Close() })

		// --- When ---
		have := CanConnect(tspy, "1s", lnr.Addr().String())

		// --- Then ---
		assert.NotNil(t, have)
		tspy.Finish()

		err := have.Close()
		assert.ErrorContain(t, "use of closed network connection", err)
	})

	t.Run("failure to connect", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFatal()
		tspy.ExpectLogContain("expected successful connection:")
		tspy.ExpectLogContain("connection refused")
		tspy.Close()

		lnr := must.Value(net.Listen("tcp", "127.0.0.1:0"))
		addr := lnr.Addr().String()
		_ = lnr.Close()

		// --- When ---
		assert.Panic(t, func() { CanConnect(tspy, "1ms", addr) })
	})

	t.Run("invalid timeout", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.ExpectFatal()
		wMsg := "expected valid time duration:\n" +
			"   have: abc\n" +
			"  error: "
		tspy.ExpectLogContain(wMsg)
		tspy.ExpectLogContain("invalid duration")
		tspy.Close()

		// --- When ---
		assert.Panic(t, func() { CanConnect(tspy, "abc", "") })
	})
}

func Test_CanNotConnect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t)
		tspy.Close()

		lnr := must.Value(net.Listen("tcp", "127.0.0.1:0"))
		addr := lnr.Addr().String()
		_ = lnr.Close()

		// --- When ---
		have := CanNotConnect(tspy, addr)

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("can connect is failure", func(t *testing.T) {
		// --- When ---
		tspy := tester.New(t)
		tspy.ExpectError()
		wMsg := "expected no connection possible:\n  address: 127.0.0.1:"
		tspy.ExpectLogContain(wMsg)
		tspy.Close()

		lnr := must.Value(net.Listen("tcp", "127.0.0.1:0"))
		t.Cleanup(func() { _ = lnr.Close() })

		have := CanNotConnect(tspy, lnr.Addr().String())

		// --- Then ---
		assert.False(t, have)
	})
}

func Test_GetLocalIP(t *testing.T) {
	// --- When ---
	ip := GetLocalIP()

	// --- Then ---
	assert.NotEmpty(t, ip.String())
	assert.False(t, ip.Equal(net.IPv4(127, 0, 0, 1)))
}
