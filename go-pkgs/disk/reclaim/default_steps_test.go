package reclaim

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRun_BrewCleanupClearsMetadataLeftovers pins the fix for a step whose
// measured footprint was not actually reclaimed: `brew cleanup -s --prune=all`
// drops downloads but leaves api/, bootsnap/ and Backup/ behind, so the step
// reported Before 52 MB / After 52 MB / Freed 0. The step now removes the
// remainder too, so Freed matches what it promised.
func TestRun_BrewCleanupClearsMetadataLeftovers(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	stubs := t.TempDir()
	// brew succeeds but deletes nothing, mimicking the leftovers.
	writeStubTool(t, stubs, "brew", "#!/bin/sh\nexit 0\n")

	cache := filepath.Join(home, "Library", "Caches", "Homebrew")
	writeFixtureFile(t, filepath.Join(cache, "api", "formula.jws.json"), 600)
	writeFixtureFile(t, filepath.Join(cache, "bootsnap", "bootsnap.compiled"), 200)

	var stdout, stderr bytes.Buffer
	res, err := Run(Options{
		Home:       home,
		Stdout:     &stdout,
		Stderr:     &stderr,
		PathEnv:    stubs,
		SkipVolume: true,
	})
	if err != nil {
		t.Fatalf("run: %v (stderr=%s)", err, stderr.String())
	}
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Fatalf("brew cache still exists after run (err=%v)", err)
	}
	if res.TotalFreed != 800 {
		t.Fatalf("TotalFreed = %d, want 800", res.TotalFreed)
	}
	var freed int64 = -1
	for _, s := range res.Steps {
		if s.ID == "brew-cleanup" {
			freed = s.Freed
		}
	}
	if freed != 800 {
		t.Fatalf("brew-cleanup Freed = %d, want 800", freed)
	}
	if strings.Contains(stderr.String(), "brew-cleanup") {
		t.Fatalf("brew-cleanup warned despite succeeding: %q", stderr.String())
	}
}

// TestRun_IOSSimNoRuntimesSkipsDelete pins the other live-run defect: with no
// runtimes installed, `xcrun simctl runtime delete all` exits 2 ("No matching
// images found to delete"), which surfaced as a spurious warning. Nothing is
// installed, so the product should not call delete at all and should stay quiet.
func TestRun_IOSSimNoRuntimesSkipsDelete(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	stubs := t.TempDir()
	writeStubTool(t, stubs, "xcrun", `#!/bin/sh
case "$*" in
  "simctl runtime list --json") echo '{}'; exit 0 ;;
esac
exit 2
`)
	// Keep the other default steps quiet so only iOS behaviour is observed.
	writeStubTool(t, stubs, "brew", "#!/bin/sh\nexit 0\n")
	writeStubTool(t, stubs, "npm", "#!/bin/sh\nexit 0\n")

	var stdout, stderr bytes.Buffer
	res, err := Run(Options{
		Home:          home,
		IncludeIOSSim: true,
		Stdout:        &stdout,
		Stderr:        &stderr,
		PathEnv:       stubs,
		SkipVolume:    true,
	})
	if err != nil {
		t.Fatalf("run: %v (stderr=%s)", err, stderr.String())
	}
	if strings.Contains(stderr.String(), "ios-sim-runtimes") {
		t.Fatalf("ios-sim-runtimes warned for a no-op delete: %q", stderr.String())
	}
	found := false
	for _, s := range res.Steps {
		if s.ID == "ios-sim-runtimes" {
			found = true
			if s.Before != 0 || s.After != 0 || s.Freed != 0 {
				t.Fatalf("ios-sim-runtimes = %+v, want all zeros", s)
			}
		}
	}
	if !found {
		t.Fatalf("ios-sim-runtimes row missing from report:\n%s", stdout.String())
	}
}
