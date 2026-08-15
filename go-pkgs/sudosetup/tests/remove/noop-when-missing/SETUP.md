# Scenario

**Feature**: Remove when nothing installed only flushes sudo cache

```
# no drop-in/manifest -> only sudo -k, no error
```

## Preconditions

- No drop-in or manifest.

## Steps

1. Leave seeds empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "noop_when_missing"
	return nil
}
```