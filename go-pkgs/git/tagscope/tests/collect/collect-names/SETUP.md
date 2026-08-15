# Scenario

**Feature**: `CollectFromNames` builds tag inventory and per-scope lineage

```
tag name list -> CollectFromNames -> CollectedTags
```

## Steps

1. Set `req.Op` to `"collect-names"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "collect-names"
	return nil
}
```