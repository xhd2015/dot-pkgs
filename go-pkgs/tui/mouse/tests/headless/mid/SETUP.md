# Scenario

**Feature**: mid-pane anchor (default pad ~8) for headless mouse geometry

```
# mid pad pushes UI into middle of 24-row PTY; ORIGIN ~ pad after CPR
fixture --anchor=mid -> pad blanks + VIEW=5 UI -> ORIGIN mid-range
```

## Steps

1. Set `req.Anchor = "mid"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Anchor = "mid"
	return nil
}
```
