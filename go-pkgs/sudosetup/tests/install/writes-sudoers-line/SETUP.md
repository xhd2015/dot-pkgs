# Scenario

**Feature**: EnsureInstalled renders sudoers line and runs visudo + install

```
# not installed -> visudo -cf temp -> install drop-in
```

## Preconditions

- No prior install.
- Runner accepts visudo and install.

## Steps

1. Leave seeds empty so install path runs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "writes_sudoers_line"
	return nil
}
```