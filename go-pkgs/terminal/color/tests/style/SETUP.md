# Scenario

**Feature**: Style wraps text with SGR when Enabled, otherwise passthrough

```
Style{Enabled} -> Green/Red/Yellow/Blue/Gray/Bold/Dim/Strike/Paint(text)
# !Enabled or empty text -> text unchanged
```

## Preconditions

- `req.Op` is `"style"`.
- Leaves set `Enabled`, `Color` (`green`|`red`|`yellow`|`gray`|`blue`|`bold`|`dim`|`strike`|`paint-gray-strike`), and `Text`.

## Steps

1. Set `req.Op` to `"style"`.
2. `Run` builds `color.Style{Enabled: req.Enabled}` and calls the named method.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "style"
	return nil
}
```
