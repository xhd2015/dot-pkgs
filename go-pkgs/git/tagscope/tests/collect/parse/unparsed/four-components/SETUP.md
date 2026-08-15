# Scenario

**Feature**: four-component versions are rejected

```
v0.0.2.1 -> ParseTagName -> ok=false
```

## Steps

1. Set `req.Name` to `v0.0.2.1`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Name = "v0.0.2.1"
	return nil
}
```