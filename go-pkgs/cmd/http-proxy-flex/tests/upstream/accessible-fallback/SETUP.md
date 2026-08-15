## Steps

- Start a TCP listener on a random port to simulate an available upstream proxy
- Run `http-proxy` pointing at that listener (default flex: health monitor always on)
- Capture initial log output, then kill

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	binPath := getBinPath(t, d)
	req.CapturedOutput = startWithLiveUpstreamAndCapture(t, binPath, "--listen-port", "19984")
	return nil
}
```
