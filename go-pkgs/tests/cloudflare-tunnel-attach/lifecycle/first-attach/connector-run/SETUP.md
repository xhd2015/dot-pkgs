# Scenario

**Feature**: first attach starts connector via runner Exec of tunnel run

```
# Attach (Runner != nil)
Attach(a.example.com)
  -> runner.Exec("cloudflared", "tunnel", …, "run", name) at least once
```

## Preconditions

- Fresh registry; first successful attach for this ConfigDir/TunnelName.
- Fake runner treats `run` as immediate success (no OS process).

## Steps

1. Attach once.
2. Assert RunCount ≥ 1 from fake call log.

## Context

- Requirement scenario 1: run count ≥ 1 on first attach.
- Matches StartSession fake-runner convention:
  `tunnel --config <path> run <name>` (alternate arg order also acceptable).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "connector-run")
	req.Domain = "a.example.com"
	req.LocalURL = "http://127.0.0.1:6321"
	req.Sequence = nil
	return nil
}
```
