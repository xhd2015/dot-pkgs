# Scenario

**Feature**: missing gh binary path returns not-found error

```
# gh binary missing
ListOwned GhBin=/nonexistent/gh -> gh not found
```

## Steps

1. Set `req.GhBin` to a path that does not exist.
2. Set `req.Owners` to `["anyuser"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = []string{"anyuser"}
	req.GhBin = "/nonexistent/path/to/gh-missing-binary"
	return nil
}```