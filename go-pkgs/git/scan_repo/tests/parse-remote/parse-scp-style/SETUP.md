# Scenario

**Feature**: SCP-style enterprise host URL parses to owner and repo

```
git@host:owner/repo.git -> ParseRemoteOwnerRepo -> owner, repo, ok=true
```

## Steps

1. Set `req.ParseURL` to `git@github.enterprise.com:acme/widget.git`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ParseURL = "git@github.enterprise.com:acme/widget.git"
	return nil
}
```