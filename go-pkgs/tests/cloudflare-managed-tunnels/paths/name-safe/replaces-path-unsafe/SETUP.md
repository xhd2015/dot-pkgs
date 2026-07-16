# Scenario

**Feature**: path-unsafe characters are stripped or replaced so result is a single path segment

```
# TunnelNameSafe
"team/shared\\x" -> no '/' or '\\' in result
```

## Preconditions

- TunnelName contains `/` and `\` path separators.

## Steps

1. Set TunnelName to a string containing both slash and backslash.

## Context

- Requirement scenario 3: result must not contain `/` (nor `\`).
- Exact replacement form is implementer-owned (`-` is preferred); assert is
  separator-free + non-empty.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "replaces-path-unsafe")
	req.TunnelName = `team/shared\x`
	return nil
}
```
