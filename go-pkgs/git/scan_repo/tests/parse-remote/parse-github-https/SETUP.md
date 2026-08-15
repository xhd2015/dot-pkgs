# Scenario

**Feature**: HTTPS GitHub URL parses to owner and repo

```
https://github.com/owner/repo.git -> ParseRemoteOwnerRepo -> owner, repo, ok=true
```

## Steps

1. Set `req.ParseURL` to `https://github.com/golang/go.git`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ParseURL = "https://github.com/golang/go.git"
	return nil
}
```