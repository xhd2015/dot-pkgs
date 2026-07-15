# Scenario

**Feature**: deep multi-segment path prefix parses like shallow subpaths

```
pkg/api/v1.0.0-dev -> ParseTagName -> PathPrefix=pkg/api/, Prerelease=dev
```

## Steps

1. Set `req.Name` to `pkg/api/v1.0.0-dev`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Name = "pkg/api/v1.0.0-dev"
	return nil
}
```