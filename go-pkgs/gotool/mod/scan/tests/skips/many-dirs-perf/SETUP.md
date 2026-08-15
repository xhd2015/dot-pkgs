# Scenario

**Feature**: performance — `Scan` over a moderate-sized repo (~100 dirs) completes within 500ms

```
# A typical Go project with many packages but only a root go.mod
# Scan must NOT spend excessive time calling git check-ignore per-directory
root + 100 subdirs -> scan.Scan -> [.] in <500ms
```

This test encodes the performance expectation reported in `wrk --dep` slowness:
`scan.Scan` on the dot-pkgs repo (~650 dirs) takes ~8.9 seconds because the
`shouldSkip` function calls `git check-ignore` as a subprocess for every non-root
directory. With ~100 directories, each subprocess call at ~15ms, the current code
spends ~1.5s overhead — far exceeding the expected 500ms budget.

When the fix is applied (removing the per-directory `CheckIgnore` call, which
`ListIgnoredDirs` already covers), this test will pass.

## Steps

1. Create a git workspace with root `go.mod`.
2. Create 100 package-level subdirectories (no go.mod in them, just placeholder files).
3. Run `Scan` and assert wall-clock time is under 500ms.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initSkipRoot(t, "example.com/root")

	// Create 100 subdirectories to simulate a realistic Go project structure.
	// Each gets a placeholder .go file so git tracks them.
	for i := 0; i < 100; i++ {
		dir := filepath.Join(ws, "pkg", string(rune('a'+i%26)), string(rune('a'+(i/26)%26)), string(rune('a'+i%10)))
		writeFile(t, filepath.Join(dir, "placeholder.go"), "package placeholder\n")
	}

	mustGit(t, ws, "add", "-A")
	mustGit(t, ws, "commit", "-m", "add packages")

	req.RootDir = ws
	return nil
}
```
