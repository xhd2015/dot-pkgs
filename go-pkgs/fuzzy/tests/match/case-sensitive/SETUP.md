# Scenario

**Feature**: case-sensitive match rejects a case-folded query

```
# Match + WithCaseSensitive
Match("aid-user", "AID") -> !OK
```

## Steps

1. Set `req.Op` to `"match"`, haystack `"aid-user"`, query `"AID"`,
   `CaseSensitive` true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match"
	req.Haystack = "aid-user"
	req.Query = "AID"
	req.CaseSensitive = true
	return nil
}
```
