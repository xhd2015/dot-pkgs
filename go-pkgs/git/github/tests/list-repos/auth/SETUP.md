# Scenario

**Feature**: authentication gate before any `ListRepos` query

```
# fail fast on auth
ListRepos -> EnsureAuthenticated -> gh api user -> login or error
```

## Steps

1. Auth leaves clear explicit owners so inference does not mask auth failures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = nil
	return nil
}
```