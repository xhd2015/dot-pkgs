# Scenario

**Feature**: `ParseRemoteOwnerRepo` parses git remote URLs without filesystem access

```
# pure URL parser
remote URL string -> ParseRemoteOwnerRepo -> owner, repo, ok
```

## Preconditions

- `req.ParseURL` is set by each leaf; `Scan` is not invoked.

## Steps

1. Clear scan-related fields so only the parser path runs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Roots = nil
	req.ListRemotes = false
	req.ListWorktrees = false
	return nil
}
```