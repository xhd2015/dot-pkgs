# Scenario

**Feature**: unauthenticated `gh api user` hints `gh auth login`

```
# auth failure
EnsureAuthenticated -> gh api user -> exit 4 -> error contains gh auth login
```

## Steps

1. Mock `gh api user` exits 4 with auth stderr.
2. Leave `req.Owners` empty (inherited nil).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.GhBin = writeAuthFailGh(t)
	return nil
}
```