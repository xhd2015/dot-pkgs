# Scenario

**Feature**: Publish sets Authorization Bearer when token is configured

```
# token present
Publisher(baseURL, token="secret-token") -> Authorization: Bearer secret-token
```

## Steps

1. Set `req.Token` to a non-empty test token.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Token = "secret-token"
	return nil
}
```
