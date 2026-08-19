# Scenario

**Feature**: MatchAll with empty tokens is OK with score 0 and one unmatched span

```
# MatchAll empty tokens
Tokens("") -> empty
MatchAll("foo", empty) -> OK score 0 one unmatched span
```

## Steps

1. Set `req.Op` to `"match_all"`, haystack `"foo"`, empty query, `Tokens`
   nil so `Run` calls `fuzzy.Tokens("")`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match_all"
	req.Haystack = "foo"
	req.Query = ""
	return nil
}
```
