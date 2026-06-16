## Preconditions

- A TCP listener can be started on a random port to simulate an upstream proxy

## Steps

- Testing `--upstream-proxy` behavior: accessible vs unreachable at startup, with and without `--fallback-direct`

```go
import "net"

func Setup(t *testing.T, req *Request) error {
	t.Log("entering upstream proxy test mode")
	return nil
}

func startWithLiveUpstreamAndCapture(t *testing.T, binPath string, extraArgs ...string) string {
	t.Helper()
	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer upstream.Close()
	upstreamAddr := upstream.Addr().String()
	args := append([]string{"--upstream-proxy", "http://" + upstreamAddr}, extraArgs...)
	return startAndCapture(t, binPath, args...)
}
```
