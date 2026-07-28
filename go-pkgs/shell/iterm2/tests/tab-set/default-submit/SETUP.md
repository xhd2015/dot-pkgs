# Scenario

**Feature**: default NoSubmit=false submits command (write text with Enter)

```
TabSpec{Command: submit-cmd} // NoSubmit omitted/false
  -> BuildTabSetNewWindowScript
  -> write text "submit-cmd"  (no "without newline" on that command line)
```

## Steps

1. One tab with command `submit-cmd`; NoSubmit left false.
2. Assert checks submit form (not only command string presence).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.TabSetName = "submit"
	req.Tabs = []TabSpecInput{
		{ID: "s1", Name: "Submit", Command: "submit-cmd", NoSubmit: false},
	}
	return nil
}
```
