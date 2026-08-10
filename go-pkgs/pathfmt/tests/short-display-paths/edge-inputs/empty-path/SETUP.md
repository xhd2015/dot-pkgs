# Scenario

**Feature**: empty path input is returned unchanged

```
# normalize
empty or Abs error -> return input unchanged
```

## Steps

1. Set `req.Path` to `""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Path = ""
	return nil
}```
