# Scenario

**Feature**: `Look(name, opts)` — ordered multi-stage binary resolution

```
Look(name, Options{injectables})
  -> direct | path | extra_dir | default_dir | candidate | login_shell:* | error
```

## Steps

1. Set `Operation=look`.
2. Child groups set the winning stage fixtures; leaves assert Path + Via.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "look"
	req.Name = "mytool"
	return nil
}
```
