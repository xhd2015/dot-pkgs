# Scenario

**Feature**: empty list dump parses to empty slice

```
"" / whitespace / # comments only
  -> ParseSessionListOutput -> [] (len 0), nil error
```

## Steps

1. Phase `parse-session-list`.
2. Primary `ListOutput` empty; Assert also covers whitespace and comments.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "parse-session-list"
	req.ListOutput = ""
	return nil
}
```
