# Scenario

**Feature**: Decode of kool's documented `\x41\x21` example is `A!`

```
Decode(`\x41\x21`) -> []byte("A!")
```

## Steps

1. Set `req.Hex` to the four-plus-four step string `\x41\x21`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Hex = `\x41\x21`
	return nil
}
```
