# Scenario

**Feature**: TunnelNameSafe produces filesystem-safe directory segments

```
# TunnelNameSafe
tunnelName -> lowercased alnum/_/- ; path-unsafe chars -> '-'
```

## Preconditions

- Mode is `name_safe`.
- Leaves supply TunnelName fixtures only.

## Steps

1. Set Mode to `name_safe`.
2. Append `name-safe` to DecisionPath.

## Context

- Prefer stable lowercasing for normal `[a-zA-Z0-9_-]` names.
- Path separators must never appear in the result.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Mode = "name_safe"
	req.DecisionPath = append(req.DecisionPath, "name-safe")
	return nil
}
```
