## Preconditions

- A TCP listener can simulate an upstream proxy
- A local HTTP server can serve as request target

## Steps

- Testing that each proxied request logs whether it goes through upstream proxy or direct

```go
import (
	"fmt"
	"net"
	"time"
)

func Setup(t *testing.T, req *Request) error {
	t.Log("entering request log test mode")
	return nil
}

func reserveTCPPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func startLocalConnectTarget(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	target := ln.Addr().String()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		conn.Close()
		ln.Close()
	}()
	return target
}

func connectThroughProxy(t *testing.T, proxyPort int, targetHost string) string {
	t.Helper()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", proxyPort), 3*time.Second)
	if err != nil {
		t.Fatalf("dial proxy: %v", err)
	}
	defer conn.Close()

	req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", targetHost, targetHost)
	if _, err := conn.Write([]byte(req)); err != nil {
		t.Fatalf("write CONNECT: %v", err)
	}

	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read CONNECT response: %v", err)
	}
	return string(buf[:n])
}
```
