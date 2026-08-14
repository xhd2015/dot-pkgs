# Scenario

**Feature**: `--name` already matches the author → nothing to change

```
RunCLI --name Alice -> Error: nothing to change
```

## Steps

1. Append `--name` `Alice` (the current author).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "--name", "Alice")
	return nil
}
```
