## Steps

- Start a TCP listener on a random port to simulate an available upstream proxy
- Run `http-proxy` pointing at that listener with `--fallback-direct`
- With `--fallback-direct`, the health check loop starts even when the initial check succeeds
- Capture initial log output, then kill

```go
func Setup(t *testing.T, req *Request) error {
	binPath := getBinPath(t)
	req.CapturedOutput = startWithLiveUpstreamAndCapture(t, binPath, "--fallback-direct", "--listen-port", "19984")
	return nil
}
```
