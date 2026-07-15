# Scenario

**Feature**: shallow subpath prerelease parses scoped prerelease fields

```
sub/v0.2.3-beta -> ParseTagName -> PathPrefix=sub/, Prerelease=beta
```

## Steps

1. Set `req.Name` to `sub/v0.2.3-beta`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Name = "sub/v0.2.3-beta"
	return nil
}
```