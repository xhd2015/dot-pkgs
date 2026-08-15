## Expected

- `resp.SelectedBudgetOK` is true and `resp.SelectedBudget == 500ms`.
- Direct checks:
  - `WalkConsumeSyncBudget(10s) == 500ms`
  - `WalkConsumeSyncBudget(30s) == 500ms`
  - `WalkConsumeSyncBudget(59s) == 500ms`
- Boundary: `10s` is the first duration of this tier (not 0).

## Errors

- `err` is nil (Run returns error only if helper is missing — then fail with
  that error).

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
	want := 500 * time.Millisecond
	if !resp.SelectedBudgetOK {
		t.Fatal("SelectedBudgetOK=false; product must export WalkConsumeSyncBudget")
	}
	if resp.SelectedBudget != want {
		t.Fatalf("SelectedBudget(30s) = %v, want %v", resp.SelectedBudget, want)
	}

	cases := []time.Duration{
		10 * time.Second,
		30 * time.Second,
		59 * time.Second,
	}
	for _, d := range cases {
		got := scan_repo.WalkConsumeSyncBudget(d)
		if got != want {
			t.Fatalf("WalkConsumeSyncBudget(%v) = %v, want %v", d, got, want)
		}
	}
}
```
