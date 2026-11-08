package reclaim

import (
	"strings"
	"testing"
)

func TestCommandFor(t *testing.T) {
	t.Parallel()
	defaults := defaultStepIDs()
	cases := []struct {
		sel    []string
		dryRun bool
		want   string
	}{
		{defaults, true, "clear-tmp --dry-run"},
		{defaults, false, "clear-tmp"},
		{append(append([]string{}, defaults...), "xcache", "dirs-cleaner"), false, "clear-tmp --include xcache,dirs-cleaner"},
		{[]string{"dirs-cleaner"}, false, "clear-tmp --only dirs-cleaner"},
		{[]string{"ms-playwright", "xcache"}, true, "clear-tmp --dry-run --only xcache,ms-playwright"},
	}
	for _, c := range cases {
		if got := CommandFor(c.sel, c.dryRun); got != c.want {
			t.Errorf("CommandFor(%v, %v) = %q, want %q", c.sel, c.dryRun, got, c.want)
		}
	}
}

func TestApplySelection(t *testing.T) {
	t.Parallel()
	if _, err := applySelection(Options{Only: []string{"xcache"}, Include: []string{"xcache"}}); err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("only+include err = %v", err)
	}
	if _, err := applySelection(Options{Only: []string{"nope"}}); err == nil || !strings.Contains(err.Error(), `unknown step "nope"`) {
		t.Fatalf("unknown err = %v", err)
	}
	o, err := applySelection(Options{Only: []string{"dirs-cleaner,xcache"}, Aggressive: true})
	if err != nil {
		t.Fatal(err)
	}
	if o.Aggressive || !o.IncludeDirsCleaner {
		t.Fatalf("only should clear aggressive and set dirs-cleaner: %+v", o)
	}
	if got := strings.Join(selectedStepIDs(o), ","); got != "xcache,dirs-cleaner" {
		t.Fatalf("selected = %s", got)
	}
}
