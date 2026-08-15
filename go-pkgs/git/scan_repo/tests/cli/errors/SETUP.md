# Scenario

**Feature**: invalid argv fails before or during scan

```
caller bad argv -> RunCLI -> stderr error, non-zero exit
```

## Preconditions

- These leaves exercise validation and flag-parse errors only.

## Steps

1. Set `req.Args` per leaf without valid scan preconditions.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Error leaves must not inherit a --root from an ancestor by accident.
	var filtered []string
	for i := 0; i < len(req.Args); i++ {
		if req.Args[i] == "--root" {
			i++
			continue
		}
		if strings.HasPrefix(req.Args[i], "--root=") {
			continue
		}
		filtered = append(filtered, req.Args[i])
	}
	req.Args = filtered
	return nil
}
```