## Preconditions

- A TCP listener can be started/stopped on a known port to simulate upstream proxy lifecycle

## Steps

- Testing dynamic upstream health monitoring: dead → available → dead transitions at runtime

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Log("entering dynamic health monitoring test mode")
	return nil
}
```
