# Scenario

**Feature**: IncludeForks false adds `--source` to gh

```
# fork filter
IncludeForks=false -> gh ... --source
```

## Steps

1. Set `req.IncludeForks` to false.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.IncludeForks = false
	return nil
}```