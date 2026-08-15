# Scenario

**Feature**: `ParsePorcelain` aggregates status line counts

```
git status --porcelain lines -> ParsePorcelain -> Counts
```

## Steps

1. Set `req.Op` to `"parse"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "parse"
	return nil
}
```
