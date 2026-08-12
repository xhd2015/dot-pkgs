# Scenario

**Feature**: `Resolve*LoginEnv` — single variable from login dump

```
Resolve{Bash,Zsh}LoginEnv(name, opts)
  -> empty name -> error
  -> hit -> value string
  -> unset/empty -> ("", nil)
  -> RunLogin fail -> error
```

## Steps

1. Set `Operation=env`.
2. Child groups set Shell; leaves set EnvName and dump / fail fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "env"
	return nil
}
```
