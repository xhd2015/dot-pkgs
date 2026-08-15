# Scenario

**Feature**: ParseRef extracts owner and repo name from github field formats

```
ParseRef("xhd2015/r") -> owner xhd2015, name r
```

## Steps

1. Set `req.ParseRefInput` to the ref string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ParseRefInput = "xhd2015/fixture-a"
	return nil
}
```