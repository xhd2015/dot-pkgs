# Scenario

**Feature**: tab commands with quotes and backslashes are escaped for write text

```
TabSpec.Command = echo "hi"\path
  -> BuildTabSetNewWindowScript
  -> write text uses EscapeCommandForAppleScript form
```

## Steps

1. One tab whose command is ``echo "hi"\x`` (double quote + backslash).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.TabSetName = "esc"
	req.Tabs = []TabSpecInput{
		{ID: "e1", Name: "Esc", Command: `echo "hi"\x`},
	}
	return nil
}
```
