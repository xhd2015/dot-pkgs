# Scenario

**Feature**: default match is case-insensitive and keeps haystack case in spans

```
# Match (default fold)
Match("aid-user-do-human-verifications", "AID") -> OK
matched span Text is "aid" (original case)
```

## Steps

1. Set `req.Op` to `"match"`, haystack
   `"aid-user-do-human-verifications"`, query `"AID"`. Leave
   `CaseSensitive` false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match"
	req.Haystack = "aid-user-do-human-verifications"
	req.Query = "AID"
	return nil
}
```
