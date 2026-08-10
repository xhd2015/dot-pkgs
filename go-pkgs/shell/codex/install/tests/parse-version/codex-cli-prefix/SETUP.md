# Scenario

**Feature**: parse `codex-cli` version line

```
"codex-cli 0.147.0" -> ParseVersion -> "0.147.0"
```

## Steps

1. Set `VersionOutput` to codex-cli style output.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.VersionOutput = "codex-cli 0.147.0"
	return nil
}
```
