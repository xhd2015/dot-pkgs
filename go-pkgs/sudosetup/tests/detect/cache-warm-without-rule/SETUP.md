# Scenario

**Feature**: detect distinguishes warm sudo cache from persistent install

```
# not installed but sudo -n true succeeds
Detect -> CacheWarm=true, Installed=false, cache-only verdict
```

## Preconditions

- No drop-in or manifest.
- Runner fakes `sudo -n true` success.

## Steps

1. Enable `SudoNTrueOK`.
2. Leave install seeds empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Action = "cache_warm_without_rule"
	req.SudoNTrueOK = true
	return nil
}
```