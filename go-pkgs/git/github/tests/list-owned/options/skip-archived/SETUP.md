# Scenario

**Feature**: default skips archived repos via `--no-archived`

```
# archived filter (default)
IncludeArchived=false -> gh ... --no-archived
```

## Steps

1. Leave `req.IncludeArchived` false (default).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.IncludeArchived = false
	return nil
}```