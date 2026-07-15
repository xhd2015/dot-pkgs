# Scenario

**Feature**: `CollectFromNames` builds tag inventory and per-scope lineage

```
tag name list -> CollectFromNames -> CollectedTags
```

## Steps

1. Set `req.Op` to `"collect-names"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "collect-names"
	return nil
}
```