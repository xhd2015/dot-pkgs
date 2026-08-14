# Scenario

**Feature**: `--name` replaces only the author name

```
RunCLI --name Bob -> author Bob <alice@example.com>; message still "fix typo"
```

## Steps

1. Append `--name` `Bob`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "--name", "Bob")
	return nil
}
```
