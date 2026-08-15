# Scenario

**Feature**: Unparseable URL returns ok=false

```
not-a-remote-url -> ParseRemoteOwnerRepo -> ok=false
```

## Steps

1. Set `req.ParseURL` to a local filesystem path (not a remote URL).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ParseURL = "/Users/me/src/local-repo"
	return nil
}
```