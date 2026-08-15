# Scenario

**Feature**: top-level `--help` prints github command usage

```
# no subcommand
RunCLI --help -> top-level usage -> stdout
```

## Steps

1. Set `req.Args` to `["--help"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
