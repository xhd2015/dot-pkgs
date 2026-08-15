# Scenario

**Feature**: StartSession A then B merges both hosts in managed state

```
# Sequence
1. StartSession a.example.com -> :7001
2. StartSession b.example.com -> :7002
   same TunnelName + ConfigDir + fake Runner
  -> state Hosts has A and B with services :7001 / :7002
```

## Preconditions

- Sequence length 2; shared TunnelName `team-shared` and ConfigDir.
- Fresh registry before step 1.
- No StopSequence.

## Steps

1. Set Sequence to A then B with distinct LocalURLs.
2. Clear single-shot Domain/LocalURL so harness uses Sequence only.
3. Clear StopSequence.

## Context

- Requirement scenario 2 / exit criteria first half.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "hosts-merge")
	req.Domain = ""
	req.LocalURL = ""
	req.Sequence = []sessionStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:7001"},
		{Domain: "b.example.com", LocalURL: "http://127.0.0.1:7002"},
	}
	req.StopSequence = nil
	req.ExpectError = false
	return nil
}
```
