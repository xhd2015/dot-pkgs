# Scenario

**Feature**: `ParseTagName` recognizes scoped semver git tags

```
tag name -> ParseTagName -> ParsedTag + ok
```

## Steps

1. Set `req.Op` to `"parse"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "parse"
	return nil
}
```