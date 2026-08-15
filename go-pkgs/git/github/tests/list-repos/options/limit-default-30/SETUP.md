# Scenario

**Feature**: zero limit defaults to 30 for `ListRepos`

```
# default limit
ListReposOptions.Limit=0 -> gh repo list ... --limit 30
```

## Steps

1. Leave `req.Limit` at 0 (unset).
2. Set explicit owner `alice` for plain owned query.
3. Mock auth and empty repo list response.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 0
	req.Owners = []string{"alice"}
	req.GhBin = writeLimitDefaultGh(t)
	return nil
}
```