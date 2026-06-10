// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package netkit_test

import (
	"fmt"
	"net"
	"strings"

	"github.com/ctx42/testkit/pkg/netkit"
)

func ExampleGetFreePort() {
	// GetFreePort asks the kernel for an unused TCP port.
	port, err := netkit.GetFreePort()
	fmt.Println(err == nil && port > 0)
	// Output:
	// true
}

func ExampleGetFreePorts() {
	// GetFreePorts returns the requested number of free ports.
	ports, _ := netkit.GetFreePorts(3)
	fmt.Println(len(ports))
	// Output:
	// 3
}

func ExampleGetLocalAddress() {
	// GetLocalAddress returns a localhost address on a free port.
	addr, _ := netkit.GetLocalAddress()

	host, _, _ := net.SplitHostPort(addr)
	fmt.Println(host)
	// Output:
	// localhost
}

func ExampleGetLocalAddress_prefix() {
	// The first argument is prepended to the address.
	addr, _ := netkit.GetLocalAddress("http://")
	fmt.Println(strings.HasPrefix(addr, "http://localhost:"))
	// Output:
	// true
}
