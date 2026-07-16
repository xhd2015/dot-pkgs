# Scenario

**Feature**: same host LocalURL change rewrites Hosts service and triggers run again

```
# Sequence
1. Attach a.example.com -> http://127.0.0.1:6321
2. Attach a.example.com -> http://127.0.0.1:6322
  -> Hosts[A].Service is :6322
  -> still exactly one host key
  -> RunCount ≥ 2
```

## Preconditions

- Sequence length 2; Domain identical; LocalURL differs.
- Fresh registry before step 1.

## Steps

1. Set Sequence for A:6321 then A:6322.
2. Clear single-shot Domain/LocalURL.

## Context

- Ingress change detection must treat service URL updates as restart-worthy.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "service-change-restart")
	req.Domain = ""
	req.LocalURL = ""
	req.Sequence = []attachStep{
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:6321"},
		{Domain: "a.example.com", LocalURL: "http://127.0.0.1:6322"},
	}
	return nil
}
```
