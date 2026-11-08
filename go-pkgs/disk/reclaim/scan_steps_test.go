package reclaim

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func testScanOptions(t *testing.T) Options {
	t.Helper()
	return Options{
		Home:       t.TempDir(),
		PathEnv:    t.TempDir(),
		SkipVolume: true,
		Stdout:     os.Stdout,
		Stderr:     os.Stderr,
	}
}

func TestScanStepsMeasuresOnlyRequestedIDs(t *testing.T) {
	opts := testScanOptions(t)
	mustWrite(t, filepath.Join(opts.Home, ".sandbox", "gocache", "a", "f"))
	mustWrite(t, filepath.Join(opts.Home, ".cache", "f"))
	mustWrite(t, filepath.Join(opts.Home, ".npm", "_cacache", "f"))

	plan, err := ScanSteps(opts, []string{"sandbox-gocache", "xcache"})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, s := range plan.Steps {
		ids = append(ids, s.ID)
	}
	if got, want := strings.Join(ids, ","), "sandbox-gocache,xcache"; got != want {
		t.Fatalf("steps = %s, want %s", got, want)
	}
	for _, s := range plan.Steps {
		if s.SizeBytes == nil || *s.SizeBytes <= 0 {
			t.Errorf("%s: size = %v, want > 0", s.ID, s.SizeBytes)
		}
		if len(s.Display) == 0 {
			t.Errorf("%s: no display ops", s.ID)
		}
		if s.Group != GroupDefault && s.Group != GroupAggressive {
			t.Errorf("%s: group = %q", s.ID, s.Group)
		}
	}
}

// ScanSteps must not touch xcrun / the unified log unless those steps are
// explicitly requested; otherwise every panel open would stall for a minute.
func TestScanStepsSkipsSystemTail(t *testing.T) {
	opts := testScanOptions(t)
	mustWrite(t, filepath.Join(opts.Home, ".cache", "f"))

	plan, err := ScanSteps(opts, []string{"xcache"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].ID != "xcache" {
		t.Fatalf("steps = %+v, want only xcache", plan.Steps)
	}
	if plan.DirsCleaner.SizeBytes != nil {
		t.Errorf("dirs_cleaner was measured despite not being requested")
	}
}

func TestScanStepsUnknownID(t *testing.T) {
	opts := testScanOptions(t)
	_, err := ScanSteps(opts, []string{"nope"})
	if err == nil || !strings.Contains(err.Error(), `unknown step "nope"`) {
		t.Fatalf("err = %v, want unknown step", err)
	}
}

func TestScanStepsEmpty(t *testing.T) {
	opts := testScanOptions(t)
	if _, err := ScanSteps(opts, nil); err == nil {
		t.Fatal("want error for empty id list")
	}
}

func TestStepInfosCoverEveryStep(t *testing.T) {
	infos := StepInfos(t.TempDir())
	seen := map[string]bool{}
	for _, in := range infos {
		if in.ID == "" || in.Summary == "" || in.Group == "" {
			t.Errorf("incomplete info: %+v", in)
		}
		if seen[in.ID] {
			t.Errorf("duplicate id %q", in.ID)
		}
		seen[in.ID] = true
	}
	for _, in := range infos {
		if len(in.Actions) == 0 {
			t.Errorf("%s: no actions", in.ID)
		}
	}
	for _, id := range knownStepIDs() {
		if !seen[id] {
			t.Errorf("StepInfos missing %q", id)
		}
	}
	if !seen["shipit"] {
		t.Error("StepInfos missing shipit")
	}
}

func TestGoModCacheIsInventoryOnly(t *testing.T) {
	home := t.TempDir()
	cache := filepath.Join(home, "gomodcache")
	mustWrite(t, filepath.Join(cache, "cache", "download", "x.zip"))
	t.Setenv("GOMODCACHE", cache)

	plan, err := ScanSteps(Options{Home: home, Stdout: io.Discard, Stderr: io.Discard}, []string{"go-mod-cache"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 1 {
		t.Fatalf("steps = %d, want 1", len(plan.Steps))
	}
	st := plan.Steps[0]
	if st.Group != GroupInventory || !st.ReadOnly {
		t.Errorf("group=%q readOnly=%v, want inventory/true", st.Group, st.ReadOnly)
	}
	if st.Command != "go clean -modcache" {
		t.Errorf("command = %q, want %q", st.Command, "go clean -modcache")
	}
	if st.SizeBytes == nil || *st.SizeBytes == 0 {
		t.Errorf("size = %v, want > 0", st.SizeBytes)
	}
	if containsString(knownStepIDs(), "go-mod-cache") {
		t.Error("go-mod-cache must stay out of the runnable step list")
	}
	if err := ValidateStepIDs("--only", []string{"go-mod-cache"}); err == nil {
		t.Error("--only go-mod-cache must be rejected")
	}
}
