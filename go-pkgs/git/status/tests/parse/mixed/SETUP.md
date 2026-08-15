# Scenario

**Feature**: mixed porcelain lines aggregate correct counts

```
2 modified + 1 untracked -> ParsePorcelain -> Modified=2, Untracked=1
```

## Steps

1. Set porcelain with two ` M` lines and one `??` line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Porcelain = " M a.txt\n M b.txt\n?? c.txt"
	return nil
}
```
