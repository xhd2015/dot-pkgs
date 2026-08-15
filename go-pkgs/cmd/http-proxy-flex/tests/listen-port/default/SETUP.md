## Steps

- Run without `--listen-port`, provide `--upstream-proxy` pointing to a dead port (default flex)
- The proxy starts, prints "listening on :7821"

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"--upstream-proxy", "http://127.0.0.1:19981",

	}
	return nil
}
```
