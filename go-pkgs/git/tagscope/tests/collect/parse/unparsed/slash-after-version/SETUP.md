# Scenario

**Feature**: extra path segments after the version are rejected

```
sub/v0.0.2/extra -> ParseTagName -> ok=false
```

## Steps

1. Set `req.Name` to `sub/v0.0.2/extra`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Name = "sub/v0.0.2/extra"
	return nil
}
```