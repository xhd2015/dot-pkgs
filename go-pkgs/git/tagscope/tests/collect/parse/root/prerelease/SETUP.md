# Scenario

**Feature**: root prerelease tag parses with suffix and non-numeric release flag

```
v0.0.2-alpha -> ParseTagName -> Prerelease=alpha, IsNumericRelease=false
```

## Steps

1. Set `req.Name` to `v0.0.2-alpha`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Name = "v0.0.2-alpha"
	return nil
}
```