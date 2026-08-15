# Scenario

**Feature**: scopes are sorted lexicographically by `VersionPrefix`

```
[z/v0.0.1, v0.0.1, a/v0.0.1] -> CollectFromNames -> Scopes a/, root, z/
```

## Steps

1. Set `req.Names` spanning three distinct scopes in non-sorted order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"z/v0.0.1", "v0.0.1", "a/v0.0.1"}
	return nil
}
```