# Scenario

**Feature**: porcelain parsing and backup status formatting

```
porcelain text -> ParsePorcelain -> Counts -> Format(FormatBackup) -> clean | dirty (...)
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/git/status` is importable.
- Pure functions; no git subprocess required.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	return nil
}
```
