# Scenario

**Feature**: `ParsePorcelain` aggregates status line counts

```
git status --porcelain lines -> ParsePorcelain -> Counts
```

## Steps

1. Set `req.Op` to `"parse"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "parse"
	return nil
}
```
