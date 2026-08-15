# Scenario

**Feature**: RenderSudoersLine for bare command without args pattern

```
# hello.sh rule -> line without trailing args
```

## Preconditions

- Rule has command only, empty `ArgsPattern`.

## Steps

1. Use default hello.sh rule from root Setup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "bare_command"
	return nil
}
```