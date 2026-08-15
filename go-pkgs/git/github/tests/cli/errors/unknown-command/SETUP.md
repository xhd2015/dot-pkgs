# Scenario

**Feature**: unrecognized top-level github command

```
# no matching subcommand
RunCLI nope -> unrecognized github command -> stderr
```

## Steps

1. Set `req.Args` to `["nope"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"nope"}
	return nil
}
```