# Scenario

**Feature**: at least one of `-m`, `--email`, `--name`, `--strip-co-author` is required

```
RunCLI abcdef -> Error: at least one of -m, --email, --name, --strip-co-author is required
```

## Steps

1. Set `req.Args` to `["abcdef"]` (SHA-shaped token, no change flag).
2. This check runs before revision resolution; no repo is created.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = []string{"abcdef"}
	return nil
}
```
