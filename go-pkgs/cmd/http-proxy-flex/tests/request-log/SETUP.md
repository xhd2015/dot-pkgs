## Preconditions

- A TCP listener can simulate an upstream proxy
- A local HTTP server can serve as request target

## Steps

- Testing that each proxied request logs whether it goes through upstream proxy or direct

```go
import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Log("entering request log test mode")
	return nil
}

// reserveTCPPort returns a free TCP port number after closing the probe
// listener. Prefer net.Listen("tcp", "127.0.0.1:0") and keeping that listener
// whenever the test needs the port bound immediately — the close→rebind gap
// races parallel doctests (EADDRINUSE).
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

func getThroughProxy(t *testing.T, proxyPort int, targetURL string) {
	t.Helper()
	proxyURL, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", proxyPort))
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
		Timeout:   3 * time.Second,
	}
	if _, err := client.Get(targetURL); err != nil {
		t.Fatalf("GET through proxy: %v", err)
	}
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
