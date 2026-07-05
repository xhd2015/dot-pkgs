# Scenario

**Feature**: wrk --projects Remote: uses shared brief branch-relation labels (plain pipe)

```
tracked upstream relation -> wrk --projects (pipe) -> Remote: identical|needs merge back|needs pull|diverged
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- No `--color`; stdout is piped (plain text, no ANSI).

## Steps

- Descendants record a tracked main repo and run `wrk --projects`.
- Assertions expect shared-vocabulary `Remote:` brief summaries.

## Context

- `(no upstream)` stays unchanged (covered elsewhere).
- Color rules are out of scope here; see `color-output/`.

```go
import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--projects"}
	req.RepoDir = req.WorkRoot
	return nil
}

func recordRemoteBriefProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	runWrkWithArgs(t, req, repoPath, "--add", repoPath)
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

func remoteBriefCompareField(t *testing.T, mainRepo, upstreamRef, currentBranch string) string {
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

func remoteBriefBranchLine(t *testing.T, repoDir string) string {
	t.Helper()
	return "Branch:       " + gitOutput(t, repoDir, "rev-parse", "--abbrev-ref", "HEAD")
}

func remoteBriefCommitLine(t *testing.T, repoDir string) string {
	t.Helper()
	short := gitOutput(t, repoDir, "rev-parse", "--short=7", "HEAD")
	subject := gitOutput(t, repoDir, "log", "-1", "--pretty=%s")
	return fmt.Sprintf("Commit:       %s  %s", short, subject)
}

func remoteBriefStatusBlockTemplate(t *testing.T, mainRepo, statusLine, remoteField, worktreesSummary string) string {
	t.Helper()
	block := fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s\n%s\nWorktrees:    %s",
		resolvePath(t, mainRepo),
		remoteBriefBranchLine(t, mainRepo),
		remoteBriefCommitLine(t, mainRepo),
		statusLine,
		remoteField,
		worktreesSummary,
	)
	return "<contains>\n" + block + "\n</contains>"
}

func initRemoteBriefRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", subject)
}

func setupRemoteBriefBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGit(t, workRoot, "init", "--bare", "-b", "main", bare)
	return bare
}

func setupRemoteBriefTrackedRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initRemoteBriefRepo(t, repo, subject)
	runGit(t, repo, "remote", "add", "origin", originBare)
	runGit(t, repo, "push", "-u", "origin", "main")
	return repo
}

func pushCommitToRemoteBriefOrigin(t *testing.T, workRoot, originBare, filename, content, subject string) {
	t.Helper()
	cloneDir := filepath.Join(workRoot, "origin-push-clone")
	runGit(t, workRoot, "clone", originBare, cloneDir)
	writeFile(t, filepath.Join(cloneDir, filename), content)
	runGit(t, cloneDir, "add", filename)
	runGit(t, cloneDir, "commit", "-m", subject)
	runGit(t, cloneDir, "push", "origin", "main")
}

func assertRemoteBriefBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	got := strings.Count(stdout, "Dir:          ")
	if got != wantBlocks {
		t.Fatalf("expected %d project blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
}

func ensureRemoteBriefHelpersUsed() {
	_ = recordRemoteBriefProject
	_ = remoteBriefFromResult
	_ = remoteBriefCompareField
	_ = remoteBriefBranchLine
	_ = remoteBriefCommitLine
	_ = remoteBriefStatusBlockTemplate
	_ = initRemoteBriefRepo
	_ = setupRemoteBriefBareOrigin
	_ = setupRemoteBriefTrackedRepo
	_ = pushCommitToRemoteBriefOrigin
	_ = assertRemoteBriefBlocksSeparated
}
```