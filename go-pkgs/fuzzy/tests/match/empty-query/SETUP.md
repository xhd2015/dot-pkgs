# Scenario

**Feature**: empty query matches with score 0 and one unmatched span

```
# Match
Match("foo", "") -> OK score 0 spans [{foo, false}]
```

## Steps

1. Set `req.Op` to `"match"`, haystack `"foo"`, empty query.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match"
	req.Haystack = "foo"
	req.Query = ""
	return nil
}
```
