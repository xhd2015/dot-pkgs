# Scenario

**Feature**: `Collect` lists tags from a live git repository

```
git tag -l in repoRoot -> Collect -> CollectedTags
```

## Steps

1. Set `req.Op` to `"collect"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "collect"
	return nil
}
```