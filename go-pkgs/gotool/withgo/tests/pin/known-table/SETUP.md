# Scenario

**Feature**: PinPatch applies the kool map; naked `1.19` matches `go1.19`

```
# every go1.14..go1.25 row plus naked 1.19
go1.19 -> go1.19.13
1.19 -> go1.19.13
go1.25 -> go1.25.0
```

## Steps

1. Set `req.PinInputs` to every kool-table input plus naked `1.19`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	cases := koolPinCases()
	req.PinInputs = make([]string, len(cases))
	for i, c := range cases {
		req.PinInputs[i] = c[0]
	}
	return nil
}
```
