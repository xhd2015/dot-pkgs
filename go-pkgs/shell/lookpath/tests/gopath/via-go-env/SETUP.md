# Scenario

**Feature**: stage 3 wins — both login shells empty/soft, then go env GOPATH

```
ResolveGoPathWith(opts)
  -> bash empty, zsh empty (soft)
  -> LookPath("go") -> goBin
  -> RunGoEnv(goBin) non-empty -> first segment
```

## Steps

1. Leaves keep bash/zsh dumps without GOPATH (or empty).
2. Inject LookPath + RunGoEnv success fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Both login shells succeed with no GOPATH.
	req.BashStdout = nulEnvDump("PATH=/usr/bin")
	req.ZshStdout = nulEnvDump("PATH=/usr/bin")
	return nil
}
```
