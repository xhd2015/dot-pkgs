# Scenario

**Feature**: stage 2 wins — bash soft-empty/fail then zsh login GOPATH hit

```
ResolveGoPathWith(opts)
  -> bash login empty|error (soft continue)
  -> ResolveZshLoginEnv("GOPATH", LoginEnv) non-empty
  -> return first segment
  # LookPath / RunGoEnv not used
```

## Steps

1. Leaves configure bash soft miss (empty or error).
2. Inject zsh dump with non-empty GOPATH.
3. Assert no go stage.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Default zsh hit path; leaves override bash soft-fail flavor.
	req.ZshStdout = nulEnvDump("GOPATH=/tmp/from-zsh")
	return nil
}
```
