# Scenario

**Feature**: all-invalid names land in `Unparsed` only

```
[release-1.0, v0.0] -> CollectFromNames -> Unparsed only
```

## Steps

1. Set `req.Names` to only unrecognized tag names.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"release-1.0", "v0.0"}
	return nil
}
```