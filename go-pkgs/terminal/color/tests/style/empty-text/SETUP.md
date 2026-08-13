# Scenario

**Feature**: enabled Style returns empty text unchanged (no escape pair)

```
Style{Enabled:true}.Green("") -> ""
```

## Steps

1. Set Enabled true, Color `"green"`, Text `""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Enabled = true
	req.Color = "green"
	req.Text = ""
	return nil
}
```
