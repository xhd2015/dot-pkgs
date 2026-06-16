## Steps

- Start a TCP listener on a random port to simulate an available upstream proxy
- Run `http-proxy` pointing at that listener
- Capture initial log output, then kill

```go
func Setup(t *testing.T, req *Request) error {
	binPath := getBinPath(t)
	req.CapturedOutput = startWithLiveUpstreamAndCapture(t, binPath, "--listen-port", "19983")
	return nil
}
```
