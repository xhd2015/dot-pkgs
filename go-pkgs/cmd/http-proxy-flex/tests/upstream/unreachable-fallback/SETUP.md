## Steps

- Run `http-proxy` with `--upstream-proxy` pointing at a dead port (default flex: fallback enabled)
- The process starts, logs fallback, then we capture output and kill

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"--upstream-proxy", "http://127.0.0.1:19985",
		"--listen-port", "19986",
	}
	return nil
}
```
