<!-- TOC -->
* [The `netkit` package](#the-netkit-package)
  * [Free ports and addresses](#free-ports-and-addresses)
    * [GetFreePort](#getfreeport)
    * [GetFreePorts](#getfreeports)
    * [GetFreePortRange](#getfreeportrange)
    * [ReservePort](#reserveport)
    * [RandomPorts](#randomports)
    * [GetLocalAddress](#getlocaladdress)
  * [Connection checks](#connection-checks)
    * [CanConnect](#canconnect)
    * [CanNotConnect](#cannotconnect)
  * [Host and Internet info](#host-and-internet-info)
    * [GetLocalIP](#getlocalip)
    * [HasNetConn / SkipOnNoNetConn](#hasnetconn--skiponnonetconn)
<!-- TOC -->

# The `netkit` package

`netkit` provides networking helpers for tests: grabbing free TCP
ports and local addresses, asserting whether a TCP address accepts
connections, and gating tests on network or Internet availability.

Most helpers return their result (and an error where one is possible)
for the caller to handle. `CanConnect`, `CanNotConnect`, and
`SkipOnNoNetConn` integrate with `tester.T`, reporting through the test
handle.

## Free ports and addresses

### GetFreePort

`GetFreePort` binds to port 0 to let the kernel allocate an unused TCP
port, then releases it and returns the number:

<!-- gmdoceg:ExampleGetFreePort -->
```go
// GetFreePort asks the kernel for an unused TCP port.
port, err := netkit.GetFreePort()
fmt.Println(err == nil && port > 0)
// Output:
// true
```

### GetFreePorts

`GetFreePorts` returns the requested number of free ports. The ports
are independently allocated and need not be consecutive:

<!-- gmdoceg:ExampleGetFreePorts -->
```go
// GetFreePorts returns the requested number of free ports.
ports, _ := netkit.GetFreePorts(3)
fmt.Println(len(ports))
// Output:
// 3
```

### GetFreePortRange

`GetFreePortRange` reserves `n` *consecutive* ports, retrying up to
`retry` times if a contiguous block cannot be found. Keep `n` small
(< 10) and `retry` modest (< 5):

```go
// Three consecutive free ports, retrying up to 5 times.
ports, err := netkit.GetFreePortRange(3, 5)
```

### ReservePort

`ReservePort` checks that a specific port can be bound; it binds and
immediately releases it, returning an error if the port is taken:

```go
// Verify a specific port is currently bindable.
err := netkit.ReservePort(8080)
```

### RandomPorts

`RandomPorts` returns `n` consecutive ports as strings, starting at a
random offset above 50000. Unlike `GetFreePortRange`, it does not
check availability:

```go
// e.g. ["54321", "54322", "54323"]
ports := netkit.RandomPorts(3)
```

### GetLocalAddress

`GetLocalAddress` returns a `localhost` address on a free port,
optionally prefixed by its first argument:

<!-- gmdoceg:ExampleGetLocalAddress -->
```go
// GetLocalAddress returns a localhost address on a free port.
addr, _ := netkit.GetLocalAddress()

host, _, _ := net.SplitHostPort(addr)
fmt.Println(host)
// Output:
// localhost
```

<!-- gmdoceg:ExampleGetLocalAddress_prefix -->
```go
// The first argument is prepended to the address.
addr, _ := netkit.GetLocalAddress("http://")
fmt.Println(strings.HasPrefix(addr, "http://localhost:"))
// Output:
// true
```

## Connection checks

### CanConnect

`CanConnect` dials the address and fails the test (via `t.Fatal`) if it
cannot connect within the timeout. On success it returns the open
connection and closes it on cleanup:

```go
// Fails the test if nothing is reachable at the address in time.
conn := netkit.CanConnect(t, "1s", "localhost:8080")
```

### CanNotConnect

`CanNotConnect` is the inverse: it returns true when the address
refuses the connection, and fails the test if something is listening:

```go
// Asserts nothing is listening on the address.
ok := netkit.CanNotConnect(t, "localhost:8080")
```

## Host and Internet info

### GetLocalIP

`GetLocalIP` returns the preferred outbound IP of the host, falling
back to loopback (`127.0.0.1`) on error:

```go
ip := netkit.GetLocalIP()
```

### HasNetConn / SkipOnNoNetConn

`HasNetConn` reports whether the Internet is reachable (checked once
and cached). `SkipOnNoNetConn` skips the test when it is not:

```go
// Skip a test that needs the Internet when it is unavailable.
netkit.SkipOnNoNetConn(t)

// Or branch on it directly.
if netkit.HasNetConn() {
    // ... online-only assertions ...
}
```
