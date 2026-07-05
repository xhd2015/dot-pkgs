# Scenario

**Feature**: wrk --projects conditional ANSI coloring via --color or TTY stdout

```
# pipe stdout (non-TTY) without --color -> plain text, aligned fields
wrk --projects -> no \x1b[ sequences

# --color forces ANSI even on pipe (doctest-safe)
wrk --projects --color -> highlight attention-worthy value portions only

# --color is global; other modes ignore it today
wrk --list --color -> git worktree list unchanged (no ANSI)
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- Color assertions use `assert.Output` v2 `<ansi-color>` tags with `--color` (pipe-safe).

## Steps

- Descendants record projects and set `req.Args` to `--projects` with or without `--color`, or `--list --color` for flag no-op.

## Context

- Red (`#31`): word `dirty`, count segments with N > 0, `Remote: diverged(...)`, worktree `N dirty` when N > 0.
- Grey (`#90`): count segments with N = 0 in dirty status lines.
- Orange (`#33`): `Remote: needs merge back(...)` and `Remote: needs pull`.
- Green (`#32`): not used on `--projects` (`clean` and `identical` stay uncolored).
- Labels (`Dir:`, `Branch:`, etc.) stay uncolored; only value substrings are wrapped.
- `Worktrees:    ` uses four spaces after the colon (aligned with other fields).

```go
import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xhd2015/gitops/git"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--projects"}
	req.RepoDir = req.WorkRoot
	return nil
}

func withProjectsColor(req *Request) {
	req.Args = []string{"--projects", "--color"}
}

func stripANSI(s string) string {
	return regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(s, "")
}

func assertNoANSI(t *testing.T, s string) {
	t.Helper()
	if strings.Contains(s, "\x1b[") {
		t.Fatalf("output must not contain ANSI escapes, got:\n%s", s)
	}
}

func recordColorProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, repoPath, "--add", repoPath)
}

func colorProjectDirLine(t *testing.T, mainRepo string) string {
	t.Helper()
	return "Dir:          " + resolvePath(t, mainRepo)
}

func colorStatusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	return "Branch:       " + gitOutput(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
}

func colorStatusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutput(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutput(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func initColorOutputRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", subject)
}

func setupColorBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGit(t, workRoot, "init", "--bare", "-b", "main", bare)
	return bare
}

func setupColorTrackedMainRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initColorOutputRepo(t, repo, subject)
	runGit(t, repo, "remote", "add", "origin", originBare)
	runGit(t, repo, "push", "-u", "origin", "main")
	return repo
}

func pushCommitToBareOrigin(t *testing.T, workRoot, originBare, filename, content, subject string) {
	t.Helper()
	cloneDir := filepath.Join(workRoot, "origin-push-clone")
	runGit(t, workRoot, "clone", originBare, cloneDir)
	writeFile(t, filepath.Join(cloneDir, filename), content)
	runGit(t, cloneDir, "add", filename)
	runGit(t, cloneDir, "commit", "-m", subject)
	runGit(t, cloneDir, "push", "origin", "main")
}

func colorCompareWithRemoteField(t *testing.T, mainRepo, upstreamRef, currentBranch string) string {
	t.Helper()
	if upstreamRef == "" {
		return "Remote:       (no upstream)"
	}
	result, err := git.CompareBranches(mainRepo, upstreamRef, currentBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, upstreamRef, currentBranch, err)
	}
	return "Remote:       " + colorRemoteBriefFromResult(result)
}

func colorRemoteBriefFromResult(result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return "identical"
	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs merge back(+%d %s)", result.CommitsAheadB, commitWord)
	case git.BranchRelationBIsAncestorOfA:
		return "needs pull"
	case git.BranchRelationDiverged:
		diverged := result.CommitsAheadA + result.CommitsAheadB
		commitWord := "commit"
		if diverged != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("diverged(%d %s)", diverged, commitWord)
	default:
		return fmt.Sprintf("unknown branch relation %v", result.Relation)
	}
}

func colorProjectStatusBlockPlain(t *testing.T, mainRepo, statusLine, remoteField, worktreesSummary string) string {
	t.Helper()
	return fmt.Sprintf("%s\n%s\n%s\nStatus:       %s\n%s\nWorktrees:    %s",
		colorProjectDirLine(t, mainRepo),
		colorStatusBranchLine(t, mainRepo),
		colorStatusCommitLine(t, mainRepo),
		statusLine,
		remoteField,
		worktreesSummary,
	)
}

func colorProjectStatusBlockTemplate(t *testing.T, mainRepo, statusLine, remoteField, worktreesSummary string) string {
	t.Helper()
	return "---\nversion: 2\n---\n" + colorProjectStatusBlockPlain(t, mainRepo, statusLine, remoteField, worktreesSummary)
}

func colorLinkedWorktreeSummary(t *testing.T, mainRepo string) string {
	t.Helper()
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		t.Fatalf("ListLinked(%q): %v", mainRepo, err)
	}
	clean, dirty := 0, 0
	for _, entry := range linked {
		counts, err := colorGitStatusCounts(t, entry.Path)
		if err != nil {
			t.Fatalf("git status counts %q: %v", entry.Path, err)
		}
		if counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0 {
			clean++
		} else {
			dirty++
		}
	}
	return fmt.Sprintf("%d total, %d dirty", clean+dirty, dirty)
}

type colorPorcelainCounts struct {
	added, changed, renamed, deleted int
}

func colorGitStatusCounts(t *testing.T, repoPath string) (colorPorcelainCounts, error) {
	t.Helper()
	out := gitOutput(t, repoPath, "status", "--porcelain")
	var counts colorPorcelainCounts
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			counts.added++
			continue
		}
		if len(line) < 2 {
			counts.changed++
			continue
		}
		x, y := line[0], line[1]
		switch {
		case x == 'R' || y == 'R':
			counts.renamed++
		case x == 'A' || y == 'A':
			counts.added++
		case x == 'D' || y == 'D':
			counts.deleted++
		default:
			counts.changed++
		}
	}
	return counts, nil
}

func addColorLinkedWorktree(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGit(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func dirtyColorWorktree(t *testing.T, wtDir, filename, content string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
}

func colorDirtyStatusSegment(n int, kind string) string {
	if n > 0 {
		return fmt.Sprintf("<ansi-color red>%d %s</ansi-color>", n, kind)
	}
	return fmt.Sprintf("<ansi-color #90>%d %s</ansi-color>", n, kind)
}

func colorFormatDirtyStatusCounts(added, changed, renamed, deleted int) string {
	return fmt.Sprintf("<ansi-color red>dirty</ansi-color> (%s, %s, %s, %s)",
		colorDirtyStatusSegment(added, "added"),
		colorDirtyStatusSegment(changed, "changed"),
		colorDirtyStatusSegment(renamed, "renamed"),
		colorDirtyStatusSegment(deleted, "deleted"),
	)
}

func colorProjectsOutputBlockCount(stdout string) int {
	return strings.Count(stdout, "Dir:          ")
}

func assertColorProjectsBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	if got := colorProjectsOutputBlockCount(stdout); got != wantBlocks {
		t.Fatalf("expected %d project blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
	if wantBlocks > 1 && !strings.Contains(stdout, "\n\n") {
		t.Fatalf("expected blank line between project blocks, got:\n%s", stdout)
	}
}

func ensureColorOutputHelpersUsed() {
	_ = withProjectsColor
	_ = stripANSI
	_ = assertNoANSI
	_ = recordColorProject
	_ = colorProjectDirLine
	_ = colorStatusBranchLine
	_ = colorStatusCommitLine
	_ = initColorOutputRepo
	_ = setupColorBareOrigin
	_ = setupColorTrackedMainRepo
	_ = pushCommitToBareOrigin
	_ = colorCompareWithRemoteField
	_ = colorRemoteBriefFromResult
	_ = colorProjectStatusBlockPlain
	_ = colorProjectStatusBlockTemplate
	_ = colorLinkedWorktreeSummary
	_ = colorGitStatusCounts
	_ = addColorLinkedWorktree
	_ = dirtyColorWorktree
	_ = colorDirtyStatusSegment
	_ = colorFormatDirtyStatusCounts
	_ = colorProjectsOutputBlockCount
	_ = assertColorProjectsBlocksSeparated
}
```