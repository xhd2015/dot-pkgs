# Scenario

**Feature**: empty Owners slice rejected

```
# no owners
ListOwned owners=[] -> at least one owner required
```

## Steps

1. Set `req.Owners` to empty slice.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = []string{}
	return nil
}```