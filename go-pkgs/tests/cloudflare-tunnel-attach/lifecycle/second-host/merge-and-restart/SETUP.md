# Scenario

**Feature**: attach B after A merges both hosts and restarts connector

```
# Sequence
1. Attach a.example.com -> :7001
2. Attach b.example.com -> :7002
  -> state has A and B
  -> config both hostnames; 404 last only
  -> RunCount after step2 > RunCount after first attach (restart)
```

## Preconditions

- Sequence length 2; shared TunnelName `team-shared` and ConfigDir.
- Fresh registry before step 1.

## Steps

1. Set Sequence to A then B with distinct LocalURLs.
2. Clear single-shot Domain/LocalURL so harness uses Sequence only.

## Context

- Harness counts all `run` Execs across both steps; assert final RunCount ≥ 2
  (first start + restart on ingress change).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "merge-and-restart")
	req.Domain = ""
	req.LocalURL = ""
	req.Sequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:7001"},
		{Domain: "b.example.com", LocalURL: "http://127.0.0.1:7002"},
	}
	return nil
}
```
