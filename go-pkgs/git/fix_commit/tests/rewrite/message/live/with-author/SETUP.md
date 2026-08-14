# Scenario

**Feature**: `-m` plus `--name` plus `--email` change message and author

```
RunCLI <sha> -m "all three" --name Bob --email bob@example.com -> new author + message
```

## Steps

1. Append `-m`, `--name`, and `--email`.
2. Committer identity and both dates stay.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "-m", "all three", "--name", "Bob", "--email", "bob@example.com")
	return nil
}
```
