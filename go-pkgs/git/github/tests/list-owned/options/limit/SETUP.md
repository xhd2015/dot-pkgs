# Scenario

**Feature**: custom Limit forwards `--limit` to gh

```
# limit flag
Options.Limit=42 -> gh repo list ... --limit 42
```

## Steps

1. Set `req.Limit` to 42.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Limit = 42
	return nil
}```