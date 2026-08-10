# Scenario

**Feature**: empty string element in names is an error

```
LookupPaths(["mytool", ""], opts) -> error (empty name)
```

## Steps

1. Set `Names` to include a non-empty name and `""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"mytool", ""}
	return nil
}
```
