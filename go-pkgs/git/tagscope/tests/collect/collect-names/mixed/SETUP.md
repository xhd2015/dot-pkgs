# Scenario

**Feature**: parsed and unparsed names are partitioned

```
[v0.0.1, release-1.0, sub/v0.0.2] -> CollectFromNames -> 2 parsed + 1 unparsed
```

## Steps

1. Set `req.Names` with two valid tags and one invalid tag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v0.0.1", "release-1.0", "sub/v0.0.2"}
	return nil
}
```