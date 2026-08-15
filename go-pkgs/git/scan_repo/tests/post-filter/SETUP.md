# Scenario

**Feature**: post-process base-path filter after optional worktree resolve (P1)

```
# walk-log foreign leak
cold seed consumer only + inject foreign parent visit in walk.jsonl
  -> warm Scan(Roots=[consumer], WarmRefreshBudget=-1, ancient last_scan_end)
  -> consume re-lists foreign parent → would discover agent-pro
  -> Result.Repos paths ⊆ under consumer; no foreign agent-pro

# ListWorktrees under root
main + linked wt both under scan root, ListWorktrees=true
  -> main.Worktrees lists linked path; no promote-only top-level from enrich

# ListWorktrees outside base
scan root = main dir; linked wt elsewhere; ListWorktrees=true
  -> outer path stripped from Worktrees (or never attached)

# ListWorktrees false
ListWorktrees=false + neighbor outside root
  -> Worktrees empty; top-level filter still drops outside paths
```

## Preconditions

- Nested root: own helpers; does not inherit parent tree `Setup`.
- Every leaf uses an explicit temp `CacheRoot` from `t.TempDir()` (never
  `$HOME/.cache/git-repo-scan`).
- Walk-log / sibling-filter leaves use fake `.git` only.
- ListWorktrees leaves require `git` on PATH (skip otherwise).
- Classic TDD: expect RED until product lands resolve-then-filter.

## Steps

1. Allocate a fresh temp `CacheRoot`.
2. Default enrichment off; `NoCache=false`; `Refresh=false`.
3. Provide path helpers, fake git fixtures, cold seed, walk-log inject, and
   real-git worktree helpers for descendants.

```go
import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Clean(abs)
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func gitAvailable(t *testing.T) bool {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
		return false
	}
	return true
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
}

// fakeGitRepo plants a minimal main-repo .git directory (objects only).
func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

func gitInitRepo(t *testing.T, dir string) {
	t.Helper()
	mkdirAll(t, dir)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
}

func gitInitialCommit(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "README"), "init\n")
	runGit(t, dir, "add", "README")
	runGit(t, dir, "commit", "-m", "init")
}

func gitWorktreeAdd(t *testing.T, mainDir, wtDir, branch string) {
	t.Helper()
	mkdirAll(t, filepath.Dir(wtDir))
	runGit(t, mainDir, "worktree", "add", "-b", branch, wtDir)
}

// coldSeedScan runs a full cold Scan that seeds home/repos.json under CacheRoot.
func coldSeedScan(t *testing.T, roots []string, cacheRoot string) {
	t.Helper()
	if cacheRoot == "" {
		t.Fatal("coldSeedScan: empty cacheRoot")
	}
	if len(roots) == 0 {
		t.Fatal("coldSeedScan: empty roots")
	}
	_, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:     roots,
		CacheRoot: cacheRoot,
		NoCache:   false,
	})
	if err != nil {
		t.Fatalf("cold seed Scan: %v", err)
	}
	_, ok, loadErr := scan_repo.LoadRepoIndex(cacheRoot, scan_repo.UniverseHome)
	if loadErr != nil {
		t.Fatalf("cold seed LoadRepoIndex: %v", loadErr)
	}
	if !ok {
		t.Fatalf("cold seed: expected home/repos.json under %s", cacheRoot)
	}
}

// injectVisitBeforeLastGenEnd inserts a visit event for path immediately
// before the last gen_end line in home/walk.jsonl, then reseals the cursor
// to the new EOF. Used so warm consume re-lists a foreign parent directory.
func injectVisitBeforeLastGenEnd(t *testing.T, cacheRoot, visitPath string) {
	t.Helper()
	logPath := scan_repo.WalkLogPath(cacheRoot)
	raw, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read walk.jsonl: %v", err)
	}
	type walkEv struct {
		Op   string `json:"op"`
		Path string `json:"path,omitempty"`
		Gen  int    `json:"gen,omitempty"`
	}
	var events []walkEv
	sc := bufio.NewScanner(bytes.NewReader(raw))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev walkEv
		if err := json.Unmarshal(line, &ev); err != nil {
			t.Fatalf("parse walk event: %v", err)
		}
		events = append(events, ev)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	// Find last gen_end index.
	lastGenIdx := -1
	for i, ev := range events {
		if ev.Op == "gen_end" {
			lastGenIdx = i
		}
	}
	if lastGenIdx < 0 {
		t.Fatal("walk.jsonl missing gen_end; cannot inject visit")
	}
	inject := walkEv{Op: "visit", Path: filepath.Clean(visitPath)}
	// Insert before last gen_end.
	newEvents := make([]walkEv, 0, len(events)+1)
	newEvents = append(newEvents, events[:lastGenIdx]...)
	newEvents = append(newEvents, inject)
	newEvents = append(newEvents, events[lastGenIdx:]...)

	var buf bytes.Buffer
	for _, ev := range newEvents {
		b, err := json.Marshal(ev)
		if err != nil {
			t.Fatal(err)
		}
		buf.Write(b)
		buf.WriteByte('\n')
	}
	if err := os.WriteFile(logPath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	// Cursor at sealed EOF so product state stays consistent; consume re-lists
	// generation visits regardless of cursor (see walk_log consume).
	if err := scan_repo.SaveWalkCursor(cacheRoot, int64(buf.Len())); err != nil {
		t.Fatalf("SaveWalkCursor: %v", err)
	}
}

// pathUnderRoot reports whether p is absRoot or a descendant.
func pathUnderRoot(absRoot, p string) bool {
	absRoot = filepath.Clean(absRoot)
	p = filepath.Clean(p)
	if p == absRoot {
		return true
	}
	sep := string(filepath.Separator)
	return len(p) > len(absRoot) && p[:len(absRoot)] == absRoot && p[len(absRoot):len(absRoot)+1] == sep
}

// resultPaths returns the set of Result repo Paths.
func resultPaths(repos []scan_repo.Repo) map[string]struct{} {
	out := make(map[string]struct{}, len(repos))
	for _, r := range repos {
		out[filepath.Clean(r.Path)] = struct{}{}
	}
	return out
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CacheRoot = t.TempDir()
	req.NoCache = false
	req.Refresh = false
	req.ListRemotes = false
	req.ListWorktrees = false
	req.WarmRefreshBudget = 0
	req.SetLastScanEnd = false
	req.SetNow = false
	// Default clocks unused until a leaf sets Set* flags.
	req.LastScanEnd = time.Time{}
	req.NowAt = time.Time{}
	return nil
}
```
