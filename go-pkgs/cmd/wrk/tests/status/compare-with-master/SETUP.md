# Scenario

**Feature**: wrk --status adds Compare with Master on linked worktrees only

```
# scan discovers main checkout and in-tree linked worktrees
wrk --status from main cwd -> scan_repo.Scan(root) -> status blocks

# linked worktree blocks compare main-repo branch vs worktree branch (kool format)
linked wt block -> Compare with Master: <kool compare-branch output>

# main checkout and nested independent repos omit the field
main / nested RepoTypeMain blocks -> no Compare with Master line
```

## Preconditions

- Git must be available.
- Linked worktrees are created inside the checkout root so `scan_repo.Scan` discovers them.
- Compare output matches `kool git compare-branch` formatting (via `git.CompareBranches`).

## Steps

- Descendants set up main repo + optional linked worktrees or nested repos, then run `wrk --status`.

## Context

- `Compare with Master:` compares the **main repo's current branch** (refA) against the **linked worktree's current branch** (refB).
- Only blocks where `worktree.IsLinked(repoPath)` is true include the field.
- Multi-line kool output keeps the label on the first line; continuation lines are indented to the label width.

```go
import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--status"}
	return nil
}

func setupMainRepoWithSubject(t *testing.T, workRoot, name, subject string) string {
	t.Helper()
	path := filepath.Join(workRoot, name)
	statusInitRepoWithSubject(t, path, subject)
	return path
}

func addLinkedWorktreeInRepo(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGit(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func commitOnMain(t *testing.T, mainRepo, filename, content, subject string) {
	t.Helper()
	writeFile(t, filepath.Join(mainRepo, filename), content)
	runGit(t, mainRepo, "add", filename)
	runGit(t, mainRepo, "commit", "-m", subject)
}

func formatKoolCompareBody(refA, refB string, result *git.CompareBranchesResult) string {
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
	case git.BranchRelationDiverged:
		commitWordA := "commit"
		if result.CommitsAheadA > 1 {
			commitWordA = "commits"
		}
		commitWordB := "commit"
		if result.CommitsAheadB > 1 {
			commitWordB = "commits"
		}
		return fmt.Sprintf("%s and %s has %d files difference\n"+
			"their most recent base is %s\n"+
			"%s has %d unique %s\n"+
			"%s has %d unique %s\n"+
			"They need to be merged",
			refA, refB, result.DiffFileCount, result.MergeBase,
			refA, result.CommitsAheadA, commitWordA,
			refB, result.CommitsAheadB, commitWordB)
	default:
		return fmt.Sprintf("unknown branch relation %v", result.Relation)
	}
}

func formatCompareField(t *testing.T, label, refA, refB string, result *git.CompareBranchesResult) string {
	t.Helper()
	body := formatKoolCompareBody(refA, refB, result)
	lines := strings.Split(body, "\n")
	out := label + lines[0]
	indent := strings.Repeat(" ", len(label))
	for _, line := range lines[1:] {
		out += "\n" + indent + line
	}
	return out
}

func compareWithMasterField(t *testing.T, mainRepo, mainBranch, wtBranch string) string {
	t.Helper()
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, mainBranch, wtBranch, err)
	}
	return formatCompareField(t, "Compare with Master: ", mainBranch, wtBranch, result)
}

func statusBlockWithCompare(t *testing.T, repoDir, relDir, statusLine, compareField string) string {
	t.Helper()
	if compareField == "" {
		return statusBlockTemplate(t, repoDir, relDir, statusLine)
	}
	return fmt.Sprintf(`<contains>
Dir:          %s
%s
%s
Status:       %s
%s
</contains>`, relDir, statusBranchLine(t, repoDir), statusCommitLine(t, repoDir), statusLine, compareField)
}

func assertNoCompareWithMaster(t *testing.T, stdout string) {
	t.Helper()
	if strings.Contains(stdout, "Compare with Master:") {
		t.Fatalf("stdout should not contain Compare with Master, got:\n%s", stdout)
	}
}

func ensureCompareWithMasterHelpersUsed() {
	_ = setupMainRepoWithSubject
	_ = addLinkedWorktreeInRepo
	_ = commitOnMain
	_ = formatKoolCompareBody
	_ = formatCompareField
	_ = compareWithMasterField
	_ = statusBlockWithCompare
	_ = assertNoCompareWithMaster
}
```