# Scenario

**Feature**: `--strip-co-author` plus `--name` both apply

```
default trailers + --strip-co-author --name Bob -> message "fix typo", author Bob
```

## Steps

1. Keep the default trailer fixture.
2. Append `--strip-co-author` `--name` `Bob`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Args = append(req.Args, "--strip-co-author", "--name", "Bob")
	return nil
}
```
