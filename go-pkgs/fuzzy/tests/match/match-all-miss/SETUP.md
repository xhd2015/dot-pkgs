# Scenario

**Feature**: MatchAll fails when any token is missing from the haystack

```
# MatchAll AND
MatchAll("followup", Tokens("aid user")) -> !OK
```

## Steps

1. Set `req.Op` to `"match_all"`, haystack `"followup"`, query `"aid user"`.
   Leave `Tokens` nil so `Run` derives them via `fuzzy.Tokens`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match_all"
	req.Haystack = "followup"
	req.Query = "aid user"
	return nil
}
```
