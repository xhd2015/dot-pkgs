# Scenario

**Feature**: unrecognized CLI flag

```
RunCLI --unknown -> flag parse error on stderr
```

## Steps

1. Pass an unknown flag without `--root`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--unknown"}
	return nil
}
```