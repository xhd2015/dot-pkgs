## Expected

- `WalkConsumeSyncBudget(60s) == 1s` and `WalkConsumeSyncBudget(2m) == 1s`.
- When `SelectedBudgetOK`, `SelectedBudget == 1s` for the injected 2m delta.
- After second Scan on the tiny fixture, last `gen_end` has `gen=2` (1s is
  enough to finish consume for this tree).

## Errors

- `err` is nil.

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	want := time.Second
	for _, d := range []time.Duration{60 * time.Second, 2 * time.Minute, time.Hour} {
		got := scan_repo.WalkConsumeSyncBudget(d)
		if got != want {
			t.Fatalf("WalkConsumeSyncBudget(%v) = %v, want %v", d, got, want)
		}
	}
	if resp.SelectedBudgetOK && resp.SelectedBudget != want {
		t.Fatalf("SelectedBudget = %v, want %v", resp.SelectedBudget, want)
	}

	if !resp.WalkLogOK {
		t.Fatal("expected walk.jsonl after consume under 1s budget")
	}
	last, ok := lastGenEnd(resp.WalkEvents)
	if !ok || last.Gen != 2 {
		t.Fatalf("expected last gen_end gen=2 under delta>=60s budget; last=%v ok=%v events=%v",
			last, ok, resp.WalkEvents)
	}
}
```
