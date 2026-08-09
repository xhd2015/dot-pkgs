# Scenario

**Feature**: `pathfmt.ShortFrom` shortens relative to an explicit baseDir

```
# ShortFrom pipeline
path + baseDir -> ShortFrom
  baseDir empty -> Getwd()
  base == home -> skip rel -> home shorten
  base != home, path under base -> rel
  else under home -> ~/...
```

## Preconditions

- Leaves set `req.Op = "short-from"`, `req.Path`, and `req.BaseDir`.
- Group does not chdir unless a leaf needs empty-base cwd behavior.

## Steps

1. Leaves configure Path, BaseDir, and Op.

## Context

- No shared temp tree at this level; leaves create their own fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "short-from"
	return nil
}
```
