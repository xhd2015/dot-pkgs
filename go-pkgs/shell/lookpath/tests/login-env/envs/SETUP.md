# Scenario

**Feature**: `Resolve*LoginEnvs` — full login environ as `[]string` KEY=value

```
Resolve{Bash,Zsh}LoginEnvs(LoginEnvOptions{RunLogin inject})
  -> parse env -0 dump -> []string like os.Environ
  -> RunLogin fail -> error
```

## Steps

1. Set `Operation=envs`.
2. Child groups set Shell; leaves set dump / fail fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "envs"
	return nil
}
```
