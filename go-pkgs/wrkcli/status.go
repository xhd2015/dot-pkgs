package wrkcli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/gitops/git"
)

type statusCounts struct {
	added   int
	changed int
	renamed int
	deleted int
}

func runStatus(workDir string, colorEnabled bool) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}

	repos, err := discoverStatusRepos(context.Background(), checkoutRoot)
	if err != nil {
		return err
	}

	for i, repo := range repos {
		if i > 0 {
			fmt.Println()
		}
		if err := printStatusBlock(checkoutRoot, repo.Path, colorEnabled); err != nil {
			return err
		}
	}
	return nil
}

func runRepos(workDir string) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	checkoutRoot, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}

	repos, err := discoverStatusRepos(context.Background(), checkoutRoot)
	if err != nil {
		return err
	}
	for _, repo := range repos {
		rel, err := filepath.Rel(checkoutRoot, repo.Path)
		if err != nil {
			return fmt.Errorf("resolve relative repo path: %w", err)
		}
		fmt.Println(filepath.ToSlash(rel))
	}
	return nil
}

func discoverStatusRepos(ctx context.Context, root string) ([]scan_repo.Repo, error) {
	return scan_repo.Scan(ctx, scan_repo.Options{Roots: []string{root}})
}

func printStatusBlock(root, repoPath string, colorEnabled bool) error {
	rel, err := filepath.Rel(root, repoPath)
	if err != nil {
		return fmt.Errorf("resolve relative repo path: %w", err)
	}
	if rel == "." {
		rel = "."
	}

	branch, err := gitOutput(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	short, err := gitOutput(repoPath, "rev-parse", "--short=7", "HEAD")
	if err != nil {
		return err
	}
	subject, err := gitOutput(repoPath, "log", "-1", "--pretty=%s")
	if err != nil {
		return err
	}
	counts, err := gitStatusCounts(repoPath)
	if err != nil {
		return err
	}

	hasMaster := worktree.IsLinked(repoPath)
	var masterBrief string
	if hasMaster {
		masterBrief, _, err = masterBriefForRepo(repoPath, branch, colorEnabled)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Dir:          %s\n", filepath.ToSlash(rel))
	fmt.Printf("Branch:       %s\n", branch)
	fmt.Printf("Commit:       %s  %s\n", short, subject)

	statusLine := formatStatusCounts(counts, colorEnabled, true)
	if hasMaster {
		fmt.Printf("Status:       %s\n", statusLine)
		fmt.Printf("Master:       %s\n", masterBrief)
	} else {
		fmt.Printf("Status:       %s\n", statusLine)
	}
	return nil
}

func projectBlockUsesColor(colorEnabled bool, counts statusCounts, remoteRelation git.BranchRelation, dirtyWorktrees, worktreeErrors int) bool {
	if !colorEnabled {
		return false
	}
	if counts.added != 0 || counts.changed != 0 || counts.renamed != 0 || counts.deleted != 0 {
		return true
	}
	if dirtyWorktrees > 0 {
		return true
	}
	if worktreeErrors > 0 {
		return true
	}
	switch remoteRelation {
	case git.BranchRelationAIsAncestorOfB, git.BranchRelationBIsAncestorOfA, git.BranchRelationDiverged:
		return true
	default:
		return false
	}
}

func statusBlockUsesColor(colorEnabled bool, counts statusCounts, hasMaster bool, masterRelation git.BranchRelation) bool {
	if !colorEnabled {
		return false
	}
	if counts.added != 0 || counts.changed != 0 || counts.renamed != 0 || counts.deleted != 0 {
		return true
	}
	if counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0 {
		return true
	}
	if hasMaster {
		switch masterRelation {
		case git.BranchRelationSame, git.BranchRelationAIsAncestorOfB, git.BranchRelationBIsAncestorOfA, git.BranchRelationDiverged:
			return true
		}
	}
	return false
}

func masterBriefForRepo(repoPath, wtBranch string, colorEnabled bool) (string, git.BranchRelation, error) {
	mainRepo, err := worktree.ReadMainRepo(repoPath)
	if err != nil {
		return "", 0, err
	}
	mainBranch, err := gitOutput(mainRepo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", 0, err
	}
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		return "", 0, err
	}
	return FormatMasterBrief(result, colorEnabled), result.Relation, nil
}

func formatCompareWithRemote(mainRepoPath, currentBranch string, colorEnabled bool) (string, error) {
	upstream, err := gitUpstreamRef(mainRepoPath)
	if err != nil {
		return "", err
	}
	if upstream == "" {
		return "Remote:       (no upstream)", nil
	}
	if err := gitFetchQuiet(mainRepoPath); err != nil {
		return "", err
	}
	result, err := git.CompareBranches(mainRepoPath, upstream, currentBranch)
	if err != nil {
		return "", err
	}
	return "Remote:       " + FormatRemoteBrief(result, colorEnabled), nil
}

func gitUpstreamRef(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "@{upstream}")
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	out, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func gitWorktreeIsClean(repoPath string) (bool, error) {
	out, err := gitCombinedOutput(repoPath, "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	counts := parseStatusCounts(strings.TrimSpace(string(out)))
	return counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0, nil
}

func gitCombinedOutput(repoPath string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	return cmd.CombinedOutput()
}

func gitCombinedOutputError(repoPath string, args ...string) string {
	out, err := gitCombinedOutput(repoPath, args...)
	if err == nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitWorktreeStatusCounts(repoPath string) (statusCounts, error) {
	clean, err := gitWorktreeIsClean(repoPath)
	if err != nil {
		return statusCounts{}, err
	}
	if clean {
		return statusCounts{}, nil
	}
	return statusCounts{changed: 1}, nil
}

func gitOutputNoOptionalLocks(repoPath string, args ...string) (string, error) {
	return gitOutput(repoPath, args...)
}

func gitOutput(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func gitStatusCounts(repoPath string) (statusCounts, error) {
	out, err := gitOutput(repoPath, "status", "--porcelain", "--untracked-files=no")
	if err != nil {
		return statusCounts{}, err
	}
	return parseStatusCounts(out), nil
}

func parseProjectStatusCounts(out string, skipUntracked map[string]struct{}) statusCounts {
	var counts statusCounts
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") && len(skipUntracked) > 0 {
			path := strings.TrimSpace(line[3:])
			path = strings.TrimSuffix(path, "/")
			if _, ok := skipUntracked[path]; ok {
				continue
			}
		}
		countStatusLine(&counts, line)
	}
	return counts
}

func parseStatusCounts(out string) statusCounts {
	var counts statusCounts
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		countStatusLine(&counts, line)
	}
	return counts
}

func gitFetchQuiet(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--quiet")
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git fetch: %w", err)
	}
	return nil
}

func gitFetchQuietNoOptionalLocks(repoPath string) error {
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--quiet")
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git fetch: %w", err)
	}
	return nil
}

func gitFetchUpstreamQuietNoOptionalLocks(repoPath, upstream string) error {
	remote, branch, ok := strings.Cut(upstream, "/")
	if !ok || remote == "" || branch == "" {
		return gitFetchQuietNoOptionalLocks(repoPath)
	}
	cmd := exec.Command("git", "-C", repoPath, "fetch", "--quiet", remote, branch)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git fetch: %w", err)
	}
	return nil
}

func countStatusLine(counts *statusCounts, line string) {
	if strings.HasPrefix(line, "??") {
		counts.added++
		return
	}
	if len(line) < 2 {
		counts.changed++
		return
	}

	x := line[0]
	y := line[1]
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

func formatStatusCounts(counts statusCounts, colorEnabled bool, greenClean bool) string {
	if counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0 {
		if colorEnabled && greenClean {
			return colorize("clean", ansiGreen)
		}
		return "clean"
	}
	if !colorEnabled {
		return fmt.Sprintf("dirty (%d added, %d changed, %d renamed, %d deleted)",
			counts.added, counts.changed, counts.renamed, counts.deleted)
	}
	return colorize("dirty", ansiRed) + " (" +
		strings.Join([]string{
			formatStatusCountSegment(counts.added, "added"),
			formatStatusCountSegment(counts.changed, "changed"),
			formatStatusCountSegment(counts.renamed, "renamed"),
			formatStatusCountSegment(counts.deleted, "deleted"),
		}, ", ") + ")"
}

func formatStatusCountSegment(n int, kind string) string {
	s := fmt.Sprintf("%d %s", n, kind)
	if n > 0 {
		return colorize(s, ansiRed)
	}
	return colorize(s, ansiGrey)
}
