# Scenario

**Feature**: EnsureInstalled aborts when visudo validation fails

```
# visudo -cf returns error -> no manifest written
```

## Preconditions

- Not installed.
- Runner fakes visudo failure.

## Steps

1. Set `VisudoFails = true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "visudo_failure"
	req.VisudoFails = true
	return nil
}
```