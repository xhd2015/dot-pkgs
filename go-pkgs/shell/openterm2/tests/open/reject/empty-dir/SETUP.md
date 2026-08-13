# Scenario

**Feature**: empty dir is rejected before any opener

```
dir="" -> OpenConfig -> error
neither opener called
```

## Steps

1. Set `Dir` to the empty string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Dir = ""
	return nil
}
```
