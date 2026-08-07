# Scenario

**Feature**: empty target uses `{Home}/Applications/iTerm.app`

```
InstallApp(extracted, "", Home=injected) -> Home/Applications/iTerm.app
```

## Steps

1. Set `UseDefaultTarget=true` so Run passes empty targetApp.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.UseDefaultTarget = true
	req.SeedExistingTarget = false
	return nil
}
```
