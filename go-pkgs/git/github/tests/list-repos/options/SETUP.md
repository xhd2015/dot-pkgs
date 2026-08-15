# Scenario

**Feature**: `ListRepos` option forwarding (limit, multi-owner)

```
# options layer
ListReposOptions Limit / Owners -> gh argv and merged results
```

## Steps

1. Clear search keywords so options leaves exercise plain owned mode unless overridden.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SearchDescription = ""
	req.SearchCode = ""
	return nil
}
```