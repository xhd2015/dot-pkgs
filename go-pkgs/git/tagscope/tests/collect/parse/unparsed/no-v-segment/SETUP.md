# Scenario

**Feature**: tags without a `v` version segment are rejected

```
release-1.0 -> ParseTagName -> ok=false
```

## Steps

1. Set `req.Name` to `release-1.0`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Name = "release-1.0"
	return nil
}
```