# Scenario

**Feature**: scrollback trim never starts the kept suffix mid escape-sequence

```
# escape-safe trim
buf longer than max, provisional cut inside ESC[?2026h
  -> trimScrollback
  -> kept suffix does not start with orphan "026h"
  -> bytes after the complete CSI remain
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` is importable.
2. Implementer provides `trimScrollback` + `TestExported_TrimScrollback`.
3. No WebSocket / full session required — pure buffer trim.

## Steps

1. Leaves build `req.Data` and `req.Max` so the provisional cut lands mid-CSI.
2. Root `Run` calls `TestExported_TrimScrollback`.
3. Assert checks prefix safety + tail marker retention.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}
```
