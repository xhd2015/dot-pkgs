# Scenario

**Feature**: subsequence match splits haystack into matched/unmatched span runs

```
# Match
Match("brainstorm", "bsm") -> OK
spans: b(matched) / rain / s(matched) / tor / m(matched)
```

## Steps

1. Set `req.Op` to `"match"`, haystack `"brainstorm"`, query `"bsm"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match"
	req.Haystack = "brainstorm"
	req.Query = "bsm"
	return nil
}
```
