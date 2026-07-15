# Scenario

**Feature**: shallow subpath numeric release parses scoped fields

```
sub/v0.2.3 -> ParseTagName -> PathPrefix=sub/, numeric release
```

## Steps

1. Set `req.Name` to `sub/v0.2.3`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Name = "sub/v0.2.3"
	return nil
}
```