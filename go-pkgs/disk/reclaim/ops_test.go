package reclaim

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStepDisplayMatchesMeta proves the typed ops table and the legacy
// stepMetas table (which the JSON `paths`/`command` fields come from) still
// describe the same action for every step: a display can never drift from
// what the step actually does.
func TestStepDisplayMatchesMeta(t *testing.T) {
	opts := Options{Home: "/Users/example"}
	ops := stepOps(opts)
	for _, def := range allSteps() {
		t.Run(def.ID, func(t *testing.T) {
			meta, ok := stepMetas[def.ID]
			if !ok {
				t.Fatalf("no meta for %q", def.ID)
			}
			cmd, paths := headlineOps(ops[def.ID])
			if meta.command != nil {
				if want := meta.command(opts.Home); cmd != want {
					t.Fatalf("command drift:\n display: %q\n legacy:  %q", cmd, want)
				}
			}
			if meta.paths != nil {
				want := meta.paths(opts.Home)
				if strings.Join(paths, "\x00") != strings.Join(want, "\x00") {
					t.Fatalf("paths drift:\n display: %q\n legacy:  %q", paths, want)
				}
			}
		})
	}
}

// TestScanEveryStepHasDisplay keeps the report UI honest: a row without a
// typed display would fall back to a bare string and lose its badge.
func TestScanEveryStepHasDisplay(t *testing.T) {
	home := t.TempDir()
	// Keep the inventory-only Go module cache inside the sandbox HOME: the
	// developer's real GOMODCACHE (GOPATH) must never be walked by tests.
	t.Setenv("GOMODCACHE", filepath.Join(home, "go", "pkg", "mod"))
	plan, err := Scan(Options{Home: home, DryRun: true, PathEnv: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, st := range plan.Steps {
		if len(st.Display) == 0 {
			t.Errorf("%s: no display op", st.ID)
			continue
		}
		if st.Command == "" && st.ReadOnly {
			continue // inventory rows may describe themselves in prose
		}
		if st.Command == "" {
			t.Errorf("%s: empty command", st.ID)
		}
	}
}

// TestRemoveOpsMatchApply is the behavioural half of the invariant: for every
// step whose headline op is a plain delete, apply it to a temp HOME and check
// that exactly the displayed targets disappear while a sibling survives.
func TestRemoveOpsMatchApply(t *testing.T) {
	cases := []struct {
		id      string
		target  string // relative to home; may contain a trailing /*
		sibling string
	}{
		{"sandbox-gocache", ".sandbox/gocache", ".sandbox/keep"},
		{"npm-npx", ".npm/_npx", ".npm/keep"},
		{"bun-cache", ".bun/install/cache", ".bun/install/keep"},
		{"xcache", ".cache", ".cachekeep"},
		{"ms-playwright", "Library/Caches/ms-playwright", "Library/Caches/keep"},
		{"opencode-snapshot", ".local/share/opencode/snapshot", ".local/share/opencode/keep"},
		{"podman-applehv-cache", ".local/share/containers/podman/machine/applehv/cache", ".local/share/containers/podman/machine/applehv/keep"},
		{"codex-logs", ".codex/logs_2.sqlite", ".codex/keep.txt"},
		{"home-logs", "ai-critic-server.log", "keep.log"},
		{"remote-agent-logs", "Library/Caches/remote-agent/a.log", "Library/Caches/remote-agent/keep.txt"},
	}
	defs := map[string]stepDef{}
	for _, def := range allSteps() {
		defs[def.ID] = def
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			home := t.TempDir()
			target := filepath.Join(home, filepath.FromSlash(tc.target))
			sibling := filepath.Join(home, filepath.FromSlash(tc.sibling))
			mustWrite(t, target)
			mustWrite(t, sibling)
			opts := Options{Home: home, PathEnv: t.TempDir(), SkipVolume: true,
				Stdout: io.Discard, Stderr: io.Discard}
			def := defs[tc.id]
			if err := def.Apply(opts); err != nil {
				t.Fatalf("apply: %v", err)
			}
			if _, err := os.Lstat(target); !os.IsNotExist(err) {
				t.Fatalf("%s still exists after apply", target)
			}
			if _, err := os.Lstat(sibling); err != nil {
				t.Fatalf("sibling %s must survive: %v", sibling, err)
			}
			if _, err := os.Lstat(filepath.Dir(sibling)); err != nil {
				t.Fatalf("sibling parent removed for %s", tc.id)
			}
		})
	}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}
