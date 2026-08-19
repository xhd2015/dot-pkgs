# Scenario

**Feature**: MatchAll ANDs whitespace-split tokens and covers the haystack

```
# MatchAll
haystack = "aid-user-do-human-verifications"
query    = "aid user"
-> OK; joinSpans == haystack; "aid" and "user" are matched spans
```

## Steps

1. Set `req.Op` to `"match_all"`, haystack
   `"aid-user-do-human-verifications"`, query `"aid user"`. Leave
   `Tokens` nil so `Run` derives them via `fuzzy.Tokens`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match_all"
	req.Haystack = "aid-user-do-human-verifications"
	req.Query = "aid user"
	return nil
}
```
