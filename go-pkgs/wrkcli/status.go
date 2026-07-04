package wrkcli

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli/storage"
	"github.com/xhd2015/gitops/git"
)

type statusCounts struct {
	added   int
	changed int
	renamed int
	deleted int
}

func runStatus(workDir string) error {
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
		if err := printStatusBlock(checkoutRoot, repo.Path); err != nil {
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

func printStatusBlock(root, repoPath string) error {
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

	fmt.Printf("Dir:          %s\n", filepath.ToSlash(rel))
	fmt.Printf("Branch:       %s\n", branch)
	fmt.Printf("Commit:       %s  %s\n", short, subject)
	fmt.Printf("Status:       %s\n", formatStatusCounts(counts))

	if worktree.IsLinked(repoPath) {
		if err := printCompareWithMaster(repoPath, branch); err != nil {
			return err
		}
	}
	return nil
}

func printProjectStatusBlock(mainRepoPath string) error {
	mainRepoPath = storage.NormalizePath(mainRepoPath)

	branch, err := gitOutput(mainRepoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	short, err := gitOutput(mainRepoPath, "rev-parse", "--short=7", "HEAD")
	if err != nil {
		return err
	}
	subject, err := gitOutput(mainRepoPath, "log", "-1", "--pretty=%s")
	if err != nil {
		return err
	}
	counts, err := gitStatusCounts(mainRepoPath)
	if err != nil {
		return err
	}

	fmt.Printf("Dir:          %s\n", mainRepoPath)
	fmt.Printf("Branch:       %s\n", branch)
	fmt.Printf("Commit:       %s  %s\n", short, subject)
	fmt.Printf("Status:       %s\n", formatStatusCounts(counts))

	if err := printCompareWithRemote(mainRepoPath, branch); err != nil {
		return err
	}

	summary, err := linkedWorktreeSummary(mainRepoPath)
	if err != nil {
		return err
	}
	fmt.Printf("Worktrees: %s\n", summary)
	return nil
}

func printCompareWithMaster(repoPath, wtBranch string) error {
	mainRepo, err := worktree.ReadMainRepo(repoPath)
	if err != nil {
		return err
	}
	mainBranch, err := gitOutput(mainRepo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		return err
	}
	fmt.Println(formatCompareField("Compare with Master: ", mainBranch, wtBranch, result))
	return nil
}

func printCompareWithRemote(mainRepoPath, currentBranch string) error {
	upstream, err := gitUpstreamRef(mainRepoPath)
	if err != nil {
		return err
	}
	if upstream == "" {
		fmt.Println("Remote:       (no upstream)")
		return nil
	}
	result, err := git.CompareBranches(mainRepoPath, upstream, currentBranch)
	if err != nil {
		return err
	}
	fmt.Printf("Remote:       %s\n", FormatRemoteBrief(result))
	return nil
}

func gitUpstreamRef(repoPath string) (string, error) {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "@{upstream}")
	out, err := cmd.Output()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

func linkedWorktreeSummary(mainRepo string) (string, error) {
	linked, err := worktree.ListLinked(mainRepo)
	if err != nil {
		return "", err
	}
	clean, dirty := 0, 0
	for _, entry := range linked {
		if worktree.IsDead(entry.Path) {
			continue
		}
		counts, err := gitWorktreeStatusCounts(entry.Path)
		if err != nil {
			return "", err
		}
		if counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0 {
			clean++
		} else {
			dirty++
		}
	}
	return fmt.Sprintf("%d total, %d dirty", clean+dirty, dirty), nil
}

func gitWorktreeStatusCounts(repoPath string) (statusCounts, error) {
	out, err := gitOutput(repoPath, "status", "--porcelain")
	if err != nil {
		return statusCounts{}, err
	}

	var counts statusCounts
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		countStatusLine(&counts, line)
	}
	return counts, nil
}

func gitOutput(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
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

	var counts statusCounts
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		countStatusLine(&counts, line)
	}
	return counts, nil
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

func formatStatusCounts(counts statusCounts) string {
	if counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0 {
		return "clean"
	}
	return fmt.Sprintf("dirty (%d added, %d changed, %d renamed, %d deleted)",
		counts.added, counts.changed, counts.renamed, counts.deleted)
}
