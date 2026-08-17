# Scenario

**Feature**: a heading-looking line inside a backtick fence is literal content

```
Fence -> Document suppresses fake boundary -> real boundary ends Section
```

## Steps

1. Provide the exact source, selector, and content needed by this case.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Source = "# Target\nbefore\n\x60\x60\x60md\n# Fake\n\x60\x60\x60\nafter\n# Next\nkeep\n"
	req.Header = "# Target"
	return nil
}
```
