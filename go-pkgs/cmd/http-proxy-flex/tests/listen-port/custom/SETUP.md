## Steps

- Run with `--listen-port 7829`, `--upstream-proxy` pointing to a dead port (default flex)
- The proxy starts, prints "listening on :7829"

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--listen-port", "7829",
		"--upstream-proxy", "http://127.0.0.1:19982",

	}
	return nil
}
```
