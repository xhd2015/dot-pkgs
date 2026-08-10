# Scenario

**Feature**: ResolveWindowSpaceIndex returns dense index for a known type-0 space id

```
index{3:0,132:1,234:2} + windowSpaceIDs[132]
  -> ResolveWindowSpaceIndex -> 1
```

## Steps

1. Phase `pure-resolve`.
2. Canonical three type-0 spaces; resolve id **132**.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "pure-resolve"
	req.Spaces = canonicalType0Spaces()
	req.ResolveSpaceIDs = []uint64{132}
	return nil
}
```
