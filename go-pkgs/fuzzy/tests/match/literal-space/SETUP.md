# Scenario

**Feature**: Match treats a space as a literal character, not a token split

```
# Match is one token
Match("aid-user", "aid user") -> !OK
```

## Steps

1. Set `req.Op` to `"match"`, haystack `"aid-user"`, query `"aid user"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "match"
	req.Haystack = "aid-user"
	req.Query = "aid user"
	return nil
}
```
