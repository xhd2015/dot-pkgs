## Steps

- Run with `--listen-port 7829`, `--upstream-proxy` pointing to a dead port, and `--fallback-direct`
- The proxy starts, prints "listening on :7829"

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--listen-port", "7829",
		"--upstream-proxy", "http://127.0.0.1:19982",
		"--fallback-direct",
	}
	return nil
}
```
