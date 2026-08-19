# Scenario

**Feature**: a consecutive hit ranks above the same letters with a gap

```
# ranking
Score Match("ab", "ab") > Score Match("a-b", "ab")
```

## Steps

1. Set `req.Op` to `"match"`, haystack `"ab"`, query `"ab"` (the consecutive
   case). Assert compares that score to `Match("a-b", "ab")`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match"
	req.Haystack = "ab"
	req.Query = "ab"
	return nil
}
```
