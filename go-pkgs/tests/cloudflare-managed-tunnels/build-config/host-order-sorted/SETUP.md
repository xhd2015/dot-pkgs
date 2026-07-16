# Scenario

**Feature**: host ingress rules are sorted by hostname; two builds are identical

```
# BuildConfigFromState determinism
Hosts map with z.* and a.* (and m.*)
  -> host rule order: a, m, z (alphabetical)
  -> second BuildConfigFromState yields same hostname sequence
```

## Preconditions

- Three hosts so order is unambiguous beyond pairwise.
- Map literal insertion order is reverse-ish; assert must not depend on it.

## Steps

1. Build StateIn with hosts keyed `z.example.com`, `a.example.com`, `m.example.com`.

## Context

- Requirement scenario 11: deterministic sorted host order.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "host-order-sorted")
	// Intentionally non-sorted map literal order.
	req.StateIn = mustState(
		"team-shared",
		"id-order",
		"/tmp/cred/order.json",
		map[string]*cloudflare.HostEntry{
			"z.example.com": {Service: "http://127.0.0.1:3"},
			"a.example.com": {Service: "http://127.0.0.1:1"},
			"m.example.com": {Service: "http://127.0.0.1:2"},
		},
	)
	return nil
}
```
