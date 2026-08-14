# Scenario

**Feature**: `--email` replaces only the author email

```
RunCLI --email bob@example.com -> Alice <bob@example.com>; message unchanged
```

## Steps

1. Append `--email` `bob@example.com`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "--email", "bob@example.com")
	return nil
}
```
