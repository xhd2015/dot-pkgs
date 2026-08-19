# Scenario

**Feature**: query with no subsequence in the haystack is not a match

```
# Match
Match("brainstorm", "zzz") -> !OK
```

## Steps

1. Set `req.Op` to `"match"`, haystack `"brainstorm"`, query `"zzz"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match"
	req.Haystack = "brainstorm"
	req.Query = "zzz"
	return nil
}
```
