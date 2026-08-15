# Scenario

**Feature**: help flag prints usage without scanning

```
caller --help -> RunCLI -> usage on stdout, exit 0
```

## Preconditions

- No `--root` required when only requesting help.

## Steps

1. Branch covers `-h` / `--help` invocation paths.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Help branch: default argv targets usage unless a leaf overrides it.
	if len(req.Args) == 0 {
		req.Args = []string{"--help"}
	}
	return nil
}
```