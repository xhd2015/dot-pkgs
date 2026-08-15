# Scenario

**Feature**: FullName is owner slash name

```
# FullName construction
owner=o name=r -> FullName o/r
```

## Steps

1. Set `req.FullNameOwner` to `o` and `req.FullNameName` to `r`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.FullNameOwner = "o"
	req.FullNameName = "r"
	return nil
}```