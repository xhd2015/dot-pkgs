# Scenario

**Feature**: `Collect` lists tags from a live git repository

```
git tag -l in repoRoot -> Collect -> CollectedTags
```

## Steps

1. Set `req.Op` to `"collect"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "collect"
	return nil
}
```