## Preconditions

- A TCP listener can be started/stopped on a known port to simulate upstream proxy lifecycle

## Steps

- Testing dynamic upstream health monitoring: dead → available → dead transitions at runtime

```go
import (
	"fmt"
	"net"
	"strings"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Log("entering dynamic health monitoring test mode")
	return nil
}

func isAddrInUse(err error) bool {
	return err != nil && strings.Contains(err.Error(), "address already in use")
}

// listenExactPort rebinds 127.0.0.1:port after a close→reserve gap. Retries
// briefly so parallel doctest probes that stole the port can release it.
func listenExactPort(port int) (net.Listener, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	var last error
	for i := 0; i < 25; i++ {
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			return ln, nil
		}
		last = err
		if !isAddrInUse(err) {
			return nil, err
		}
		time.Sleep(40 * time.Millisecond)
	}
	return nil, last
}

// withAddrInUseRetry re-runs fn when a parallel leaf stole a reserved port.
func withAddrInUseRetry(fn func() error) error {
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		last = fn()
		if last == nil || !isAddrInUse(last) {
			return last
		}
	}
	return last
}

func reserveClosedPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port, nil
}
```
