# Scenario

**Feature**: first Attach into an empty managed registry for one hostname

```
# first attach (empty state)
Attach(A)
  -> Hosts = {A}
  -> config.yml: host rule + http_status:404
  -> DNS route for A
  -> connector run ≥ 1
```

## Preconditions

- Managed tunnel directory has no prior state (fresh ConfigDir).
- Single Attach call (Sequence empty).
- Shared fixtures: domain `a.example.com`, service `http://127.0.0.1:6321`
  unless a leaf overrides.

## Steps

1. Append `first-attach` to DecisionPath.
2. Set Domain / LocalURL fixtures when unset.
3. Leaves assert one observable facet (registry/config, DNS, or run count).

## Context

- Scenarios 1, 5, 6 split across first-attach leaves (MECE by
  observation, same mutation).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "first-attach")
	if req.Domain == "" {
		req.Domain = "a.example.com"
	}
	if req.LocalURL == "" {
		req.LocalURL = "http://127.0.0.1:6321"
	}
	return nil
}
```
