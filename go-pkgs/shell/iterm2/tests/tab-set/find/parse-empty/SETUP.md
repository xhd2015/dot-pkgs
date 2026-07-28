# Scenario

**Feature**: empty find output parses to an empty session list

```
"" / whitespace -> ParseTabSetFindOutput -> [] (len 0), nil error
```

## Steps

1. Phase `parse-find`.
2. `FindOutput` is empty string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.Phase = "parse-find"
	req.FindOutput = ""
	return nil
}
```
