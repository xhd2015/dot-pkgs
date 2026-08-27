# shell/codex/usage — Codex account usage via shell/codex/api

## Version

0.0.1

Classic TDD doctests for package
`github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/usage`.

**Default layer: L2** in-process library API with injectable `GetJSON` / `LoadAuth`.
No real network, no real `~/.codex/auth.json`. Parallel-safe.

**Out of scope:** token refresh writeback, app-server JSON-RPC, Marcus wiring,
live ChatGPT e2e.

Fixtures use generic identities (`user@example.com`,
`00000000-0000-4000-8000-000000000001`) — no personal or employer hosts in
scenario copy beyond the public ChatGPT product URLs used as API constants.

## DSN (Domain Specific Notion)

### Participants

- **`Fetch(ctx, opts)`** — load auth (file or injected `Auth`) → GET
  `api.CodexUsageURL` → on failure GET `api.WhamUsageURL` → `NormalizeJSON`.
- **`NormalizeJSON(raw, source)`** — map `spend_control.individual_limit` or
  `rate_limit.primary_window` into `Snapshot`.
- **`Snapshot`** — `PlanType`, `RemainingPercent`, `UsedPercent`, `ResetAt`,
  `Source` (`codex/usage` | `wham/usage`), `Email`.

### Behaviors

**Fetch**

- Codex endpoint OK → `Source=codex/usage`, remaining/used from spend_control
- Codex fails, WHAM OK → `Source=wham/usage`
- Both fail → error

**Normalize**

- Covered primarily by L1 unit tests; L2 leaves exercise Fetch end-to-end.

## Tree

```
fetch/
├── codex-ok/          # Codex usage endpoint succeeds
├── wham-fallback/     # Codex fails → WHAM succeeds
└── both-fail/         # both endpoints fail → error
```

## How to Run

```sh
# from go-pkgs module root
doctest vet ./shell/codex/usage/tests
doctest test ./shell/codex/usage/tests
```

## Types + Run + helpers

```go
import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/api"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/codex/usage"
)

const fixtureSpendJSON = `{
  "plan_type": "business",
  "email": "user@example.com",
  "rate_limit": null,
  "spend_control": {
    "individual_limit": {
      "used_percent": 62,
      "remaining_percent": 38,
      "reset_at": 1788220800
    }
  }
}`

// Request is filled root→leaf.
type Request struct {
	Operation string // fetch

	// FetchMode: codex-ok | wham-fallback | both-fail
	FetchMode string
}

// Response observes Fetch outputs.
type Response struct {
	PlanType         string
	RemainingPercent int
	UsedPercent      int
	Source           string
	Email            string
	ResetUnix        int64
	GetURLs          []string
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error")
	}
}

func assertEqual(t *testing.T, name string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %#v want %#v", name, got, want)
	}
}

func assertContains(t *testing.T, name, got, substr string) {
	t.Helper()
	if !strings.Contains(got, substr) {
		t.Fatalf("%s: %q does not contain %q", name, got, substr)
	}
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	if req.Operation == "" {
		t.Fatal("Operation not set")
	}
	if req.Operation != "fetch" {
		t.Fatalf("unknown Operation %q", req.Operation)
	}

	var urls []string
	get := func(ctx context.Context, opts api.GetOpts) ([]byte, error) {
		urls = append(urls, opts.URL)
		switch req.FetchMode {
		case "codex-ok":
			if strings.Contains(opts.URL, "codex/usage") {
				return []byte(fixtureSpendJSON), nil
			}
			return nil, fmt.Errorf("unexpected url %q", opts.URL)
		case "wham-fallback":
			if strings.Contains(opts.URL, "codex/usage") {
				return nil, fmt.Errorf("codex api: HTTP 403 from usage endpoint")
			}
			if strings.Contains(opts.URL, "wham/usage") {
				return []byte(fixtureSpendJSON), nil
			}
			return nil, fmt.Errorf("unexpected url %q", opts.URL)
		case "both-fail":
			return nil, fmt.Errorf("codex api: HTTP 403 from usage endpoint")
		default:
			return nil, fmt.Errorf("unknown FetchMode %q", req.FetchMode)
		}
	}

	snap, err := usage.Fetch(context.Background(), usage.FetchOpts{
		Auth: api.Auth{
			AccessToken: "access-token-fixture",
			AccountID:   "00000000-0000-4000-8000-000000000001",
		},
		GetJSON: get,
	})
	resp := &Response{GetURLs: urls}
	if err != nil {
		return resp, err
	}
	resp.PlanType = snap.PlanType
	resp.RemainingPercent = snap.RemainingPercent
	resp.UsedPercent = snap.UsedPercent
	resp.Source = snap.Source
	resp.Email = snap.Email
	if !snap.ResetAt.IsZero() {
		resp.ResetUnix = snap.ResetAt.Unix()
	}
	return resp, nil
}
```
