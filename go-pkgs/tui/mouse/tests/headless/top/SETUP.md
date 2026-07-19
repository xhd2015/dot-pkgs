# Scenario

**Feature**: top anchor (pad 0) for headless mouse geometry

```
# top: no pad; ORIGIN near 0 after CPR; click still maps to btn-a
fixture --anchor=top -> ORIGIN≈0 -> click localY=3 -> HIT btn-a
```

## Steps

1. Set `req.Anchor = "top"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Anchor = "top"
	return nil
}
```
