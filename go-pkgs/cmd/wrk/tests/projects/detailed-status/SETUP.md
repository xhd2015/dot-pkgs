# Scenario

**Feature**: wrk --projects prints detailed status blocks per recorded main repo

```
# each recorded project path renders a full status block (absolute Dir)
wrk --projects -> status block per project (lexicographic order)

# extra fields vs wrk --status on main repo
Remote: <brief upstream sync summary>
Worktrees:    N total, M dirty  (linked worktrees only, always shown; four spaces after colon)
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- `wrk --projects` is standalone; empty `projects.json` yields exit 0 and empty stdout.

## Steps

- Descendants record projects via `wrk --add` or auto-record, then run `wrk --projects`.

## Context

- `Dir` is the **absolute** normalized main-repo path.
- `Remote:` uses brief sync summary from `CompareBranches(mainRepo, upstreamRef, currentBranch)`; no upstream → `(no upstream)`.
- `Worktrees:` counts linked worktrees only (`worktree.ListLinked`); clean when all porcelain counts are zero.
- Blocks are separated by a blank line; project order is lexicographic by absolute path.

```go
import (
	"fmt"
	"path/filepath"
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

func statusBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	branch := gitOutput(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	return "Branch:       " + branch
}

func statusCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutput(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutput(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func recordProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, repoPath, "--add", repoPath)
}

func projectDirLine(t *testing.T, mainRepo string) string {
	t.Helper()
	return "Dir:          " + resolvePath(t, mainRepo)
}

func formatCompareRemoteField(t *testing.T, label, upstreamRef, currentBranch string, result *git.CompareBranchesResult) string {
	t.Helper()
	body := formatKoolCompareBodyFull(upstreamRef, currentBranch, result)
	lines := strings.Split(body, "\n")
	out := label + lines[0]
	indent := strings.Repeat(" ", len(label))
	for _, line := range lines[1:] {
		out += "\n" + indent + line
	}
	return out
}

func compareWithRemoteField(t *testing.T, mainRepo, upstreamRef, currentBranch string) string {
	t.Helper()
	if upstreamRef == "" {
		return "Remote:       (no upstream)"
	}
	result, err := git.CompareBranches(mainRepo, upstreamRef, currentBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, upstreamRef, currentBranch, err)
	}
	return "Remote:       " + remoteBriefFromResult(result)
}

func remoteBriefFromResult(result *git.CompareBranchesResult) string {
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

func projectStatusBlockExact(t *testing.T, mainRepo, statusLine, compareRemoteField, worktreesSummary string) string {
	t.Helper()
	return fmt.Sprintf("%s\n%s\n%s\nStatus:       %s\n%s\nWorktrees:    %s",
		projectDirLine(t, mainRepo),
		statusBranchLine(t, mainRepo),
		statusCommitLine(t, mainRepo),
		statusLine,
		compareRemoteField,
		worktreesSummary,
	)
}

func projectStatusBlockTemplate(t *testing.T, mainRepo, statusLine, compareRemoteField, worktreesSummary string) string {
	t.Helper()
	return "<contains>\n" + projectStatusBlockExact(t, mainRepo, statusLine, compareRemoteField, worktreesSummary) + "\n</contains>"
}

func initDetailedStatusRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", subject)
}

func setupBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGit(t, workRoot, "init", "--bare", "-b", "main", bare)
	return bare
}

func setupTrackedMainRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initDetailedStatusRepo(t, repo, subject)
	runGit(t, repo, "remote", "add", "origin", originBare)
	runGit(t, repo, "push", "-u", "origin", "main")
	return repo
}

func addLinkedWorktreeForProject(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGit(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func dirtyWorktree(t *testing.T, wtDir, filename, content string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
}

func linkedWorktreeSummary(t *testing.T, mainRepo string) string {
	t.Helper()
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		t.Fatalf("ListLinked(%q): %v", mainRepo, err)
	}
	clean, dirty := 0, 0
	for _, entry := range linked {
		counts, err := gitStatusCountsForRepo(t, entry.Path)
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

type porcelainCounts struct {
	added, changed, renamed, deleted int
}

func gitStatusCountsForRepo(t *testing.T, repoPath string) (porcelainCounts, error) {
	t.Helper()
	out := gitOutput(t, repoPath, "status", "--porcelain")
	var counts porcelainCounts
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

func formatKoolCompareBodyFull(refA, refB string, result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return fmt.Sprintf("%s and %s are identical", refA, refB)
	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("%s is newer(%s +%d %s -> %s)\nto fast forward, on %s: \n   git merge --ff-only  %s",
			refB, refA, result.CommitsAheadB, commitWord, refB, refA, refB)
	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("%s is newer(%s +%d %s -> %s)\nto fast forward, on %s: \n   git merge --ff-only  %s",
			refA, refB, result.CommitsAheadA, commitWord, refA, refB, refA)
	default:
		return fmt.Sprintf("%s and %s diverged", refA, refB)
	}
}

func projectsOutputBlockCount(stdout string) int {
	return strings.Count(stdout, "Dir:          ")
}

func assertProjectsBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	if got := projectsOutputBlockCount(stdout); got != wantBlocks {
		t.Fatalf("expected %d project blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
	if wantBlocks > 1 && !strings.Contains(stdout, "\n\n") {
		t.Fatalf("expected blank line between project blocks, got:\n%s", stdout)
	}
}

func ensureDetailedStatusHelpersUsed() {
	_ = recordProject
	_ = projectDirLine
	_ = projectStatusBlockExact
	_ = projectStatusBlockTemplate
	_ = compareWithRemoteField
	_ = initDetailedStatusRepo
	_ = setupBareOrigin
	_ = setupTrackedMainRepo
	_ = addLinkedWorktreeForProject
	_ = dirtyWorktree
	_ = linkedWorktreeSummary
	_ = gitStatusCountsForRepo
	_ = formatCompareRemoteField
	_ = formatKoolCompareBodyFull
	_ = projectsOutputBlockCount
	_ = assertProjectsBlocksSeparated
}
```