## Steps

- The `--help` flag was set in upstream SETUP.md, no additional args needed for this leaf

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Log("help show: verifying --help flag output")
	return nil
}
```
