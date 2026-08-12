# Scenario

**Feature**: stage 4 — all prior stages soft-empty/fail → filepath.Join(Home, "go")

```
ResolveGoPathWith(opts)
  -> bash/zsh empty|error (soft)
  -> LookPath miss | RunGoEnv error|empty (soft)
  -> filepath.Join(Home, "go")
```

## Steps

1. Leaves force soft-fail combinations for go stage.
2. Inject Home (root already sets Home under WorkDir).
3. Assert Path equals `filepath.Join(Home, "go")`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Login shells succeed without GOPATH so cascade reaches go / home.
	req.BashStdout = nulEnvDump("PATH=/usr/bin")
	req.ZshStdout = nulEnvDump("PATH=/usr/bin")
	return nil
}
```
