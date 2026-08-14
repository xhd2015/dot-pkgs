# Scenario

**Feature**: `--name` and `--email` together; message unchanged

```
RunCLI --name Bob --email bob@example.com -> Bob <bob@example.com>; message "fix typo"
```

## Steps

1. Append `--name` `Bob` `--email` `bob@example.com`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "--name", "Bob", "--email", "bob@example.com")
	return nil
}
```
