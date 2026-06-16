## Steps

- Run without `--listen-port`, provide `--upstream-proxy` pointing to a dead port and `--fallback-direct`
- The proxy starts, prints "listening on :7821"

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--upstream-proxy", "http://127.0.0.1:19981",
		"--fallback-direct",
	}
	return nil
}
```
