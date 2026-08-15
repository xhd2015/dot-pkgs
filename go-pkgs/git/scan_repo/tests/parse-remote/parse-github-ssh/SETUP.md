# Scenario

**Feature**: SSH-style GitHub URL parses to owner and repo

```
git@github.com:owner/repo.git -> ParseRemoteOwnerRepo -> owner, repo, ok=true
```

## Steps

1. Set `req.ParseURL` to `git@github.com:xhd2015/lifelog.git`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ParseURL = "git@github.com:xhd2015/lifelog.git"
	return nil
}
```