# Scenario

**Feature**: two-component versions are rejected

```
v0.0 -> ParseTagName -> ok=false
```

## Steps

1. Set `req.Name` to `v0.0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Name = "v0.0"
	return nil
}
```