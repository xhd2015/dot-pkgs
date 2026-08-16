# Scenario

**Feature**: `ShortEnv` is a thin wrapper over `ShortEnvFrom` with process env

```
# current-env branch
path -> ShortEnv -> ShortEnvFrom(path, os.Environ())
```

## Preconditions

- Does not assert host-specific `$VAR` values (machine env is unknown).
- Only proves wrapper equality with injectable `ShortEnvFrom` + live environ.

## Steps

1. Set `req.Op = "current"`.
2. Leaf sets a concrete `req.Path` (env field unused).

## Context

- Parallel-safe: read `os.Environ()` in Assert only; never Setenv.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "current"
	return nil
}
```
