# Scenario

**Feature**: escape double quotes and backslashes in titles

```
# input: say "hi"\x
Escape… -> say \"hi\"\\x
```

## Steps

1. EscapeInput = `say "hi"\x`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.EscapeInput = `say "hi"\x`
	return nil
}
```
