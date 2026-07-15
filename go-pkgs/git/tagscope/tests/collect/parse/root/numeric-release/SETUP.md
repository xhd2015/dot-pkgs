# Scenario

**Feature**: root numeric release tag parses as numeric release

```
v0.0.1 -> ParseTagName -> numeric release at root scope
```

## Steps

1. Set `req.Name` to `v0.0.1`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Name = "v0.0.1"
	return nil
}
```