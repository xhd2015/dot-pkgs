# Scenario

**Feature**: two hosts both appear in ingress; catch-all 404 is the final rule only

```
# BuildConfigFromState
hosts: b.example.com, a.example.com
  -> both hostnames present in Ingress
  -> last rule only: empty hostname + http_status:404
```

## Preconditions

- StateIn has exactly two Hosts entries with distinct services.

## Steps

1. Build StateIn with two hosts (map insertion order not relied upon).

## Context

- Requirement scenarios 9 and 10 (multi-host + 404 last).

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/cloudflare"
)

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.DecisionPath = append(req.DecisionPath, "multi-host-both-and-404-last")
	req.StateIn = mustState(
		"team-shared",
		"id-multi-host",
		"/tmp/cred/multi.json",
		map[string]*cloudflare.HostEntry{
			"b.example.com": {Service: "http://127.0.0.1:7002"},
			"a.example.com": {Service: "http://127.0.0.1:7001"},
		},
	)
	return nil
}
```
