# Scenario

**Feature**: porcelain parsing and backup status formatting

```
porcelain text -> ParsePorcelain -> Counts -> Format(FormatBackup) -> clean | dirty (...)
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/status` is importable.
- Pure functions; no git subprocess required.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}
```
