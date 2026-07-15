# Scenario

**Feature**: IncrementTag bumps trailing numeric segment of release tag

```
release tag name -> IncrementTag -> next release tag name
```

## Steps

1. Set `req.Op` to `"increment"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "increment"
	return nil
}
```