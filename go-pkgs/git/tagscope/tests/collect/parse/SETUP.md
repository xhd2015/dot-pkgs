# Scenario

**Feature**: `ParseTagName` recognizes scoped semver git tags

```
tag name -> ParseTagName -> ParsedTag + ok
```

## Steps

1. Set `req.Op` to `"parse"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "parse"
	return nil
}
```