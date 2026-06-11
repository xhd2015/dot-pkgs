## Steps

- Start a TCP listener on a random port to simulate an available upstream proxy
- Run `http-proxy` pointing at that listener with `--fallback-direct`
- With `--fallback-direct`, the health check loop starts even when the initial check succeeds
- Capture initial log output, then kill

```go
func Run(t *testing.T, req *Request) (*Response, error) {
	binPath := getBinPath(t)
	output := startWithLiveUpstreamAndCapture(t, binPath, "--fallback-direct")
	return &Response{Output: output, ExitCode: 0}, nil
}
```
