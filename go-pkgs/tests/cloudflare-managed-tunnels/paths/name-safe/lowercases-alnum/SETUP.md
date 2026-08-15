# Scenario

**Feature**: mixed-case hyphenated name lowercases to stable safe segment

```
# TunnelNameSafe
"My-Tunnel" -> "my-tunnel"
```

## Preconditions

- Name uses only letters, digits, hyphen — no path separators.

## Steps

1. Set TunnelName to `My-Tunnel`.

## Context

- Requirement scenario 2: locked expected `my-tunnel`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "lowercases-alnum")
	req.TunnelName = "My-Tunnel"
	return nil
}
```
