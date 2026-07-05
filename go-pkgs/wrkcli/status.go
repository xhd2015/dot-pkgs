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
	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli/storage"
	"github.com/xhd2015/gitops/git"
)

type statusCounts struct {
	added   int
	changed int
	renamed int
	deleted int
}

type statusBlockPrintOpts struct {
	forceRel   string
	showMaster *bool
}

func runStatus(workDir string, colorEnabled bool, fetchEnabled bool) error {
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

	if mainRepo, ok := linkedInTreeMainRepo(cwd); ok {
		return runStatusLinkedInTreeCwd(cwd, mainRepo, colorEnabled)
	}

	repos, err := discoverStatusRepos(context.Background(), checkoutRoot)
	if err != nil {
		return err
	}

	scanPaths := make(map[string]struct{}, len(repos))
	for _, repo := range repos {
		scanPaths[storage.NormalizePath(repo.Path)] = struct{}{}
	}

	var appendEntries []worktree.Entry
	if worktree.IsMainRepo(checkoutRoot) {
		linked, err := worktree.ListLinked(checkoutRoot)
		if err != nil {
			return err
		}
		for _, entry := range linked {
			if _, ok := scanPaths[storage.NormalizePath(entry.Path)]; ok {
				continue
			}
			appendEntries = append(appendEntries, entry)
		}
	}

	scanColorEnabled := colorEnabled && len(appendEntries) == 0
	showRemote := worktree.IsMainRepo(checkoutRoot)
	effectiveFetch := fetchEnabled && showRemote

	blocksPrinted := 0
	for _, repo := range repos {
		if blocksPrinted > 0 {
			fmt.Println()
		}
		if err := printStatusBlock(checkoutRoot, repo.Path, scanColorEnabled, showRemote, effectiveFetch, statusBlockPrintOpts{}); err != nil {
			return err
		}
		blocksPrinted++
	}

	for _, entry := range appendEntries {
		if blocksPrinted > 0 {
			fmt.Println()
		}
		printAppendedLinkedBlock(checkoutRoot, entry.Path, colorEnabled)
		blocksPrinted++
	}
	return nil
}

func linkedInTreeMainRepo(cwd string) (string, bool) {
	if !worktree.IsLinked(cwd) {
		return "", false
	}
	mainRepo, err := worktree.ReadMainRepo(cwd)
	if err != nil {
		return "", false
	}
	cleanMain := filepath.Clean(mainRepo)
	cleanCwd := filepath.Clean(cwd)
	if cleanCwd == cleanMain {
		return "", false
	}
	rel, err := filepath.Rel(cleanMain, cleanCwd)
	if err != nil {
		return "", false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return mainRepo, true
}

func runStatusLinkedInTreeCwd(cwd, mainRepo string, colorEnabled bool) error {
	repos, err := discoverStatusRepos(context.Background(), mainRepo)
	if err != nil {
		return err
	}

	blocksPrinted := 0
	printBlock := func(repoPath string, opts statusBlockPrintOpts) error {
		if blocksPrinted > 0 {
			fmt.Println()
		}
		blocksPrinted++
		return printStatusBlock(mainRepo, repoPath, colorEnabled, false, false, opts)
	}

	showMasterFalse := false
	if err := printBlock(cwd, statusBlockPrintOpts{forceRel: ".", showMaster: &showMasterFalse}); err != nil {
		return err
	}

	showMasterTrue := true
	for _, repo := range repos {
		if worktree.IsMainRepo(repo.Path) || !worktree.IsLinked(repo.Path) {
			continue
		}
		if err := printBlock(repo.Path, statusBlockPrintOpts{showMaster: &showMasterTrue}); err != nil {
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

func printAppendedLinkedBlock(mainRepo, repoPath string, colorEnabled bool) {
	dirLine := storage.NormalizePath(repoPath)

	if worktree.IsDead(repoPath) {
		fmt.Printf("Dir:          %s\n", dirLine)
		fmt.Printf("Status:       prunable\n")
		return
	}

	branch, err := gitOutput(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		printAppendedBrokenBlock(dirLine, gitCombinedOutputError(repoPath, "status", "--porcelain"), colorEnabled)
		return
	}
	short, err := gitOutput(repoPath, "rev-parse", "--short=7", "HEAD")
	if err != nil {
		printAppendedBrokenBlock(dirLine, gitCombinedOutputError(repoPath, "status", "--porcelain"), colorEnabled)
		return
	}
	subject, err := gitOutput(repoPath, "log", "-1", "--pretty=%s")
	if err != nil {
		printAppendedBrokenBlock(dirLine, gitCombinedOutputError(repoPath, "status", "--porcelain"), colorEnabled)
		return
	}
	counts, err := gitStatusCounts(repoPath)
	if err != nil {
		printAppendedBrokenBlock(dirLine, gitCombinedOutputError(repoPath, "status", "--porcelain"), colorEnabled)
		return
	}
	masterBrief, _, err := masterBriefForRepo(repoPath, branch, colorEnabled)
	if err != nil {
		printAppendedBrokenBlock(dirLine, gitCombinedOutputError(repoPath, "status", "--porcelain"), colorEnabled)
		return
	}

	fmt.Printf("Dir:          %s\n", dirLine)
	fmt.Printf("Branch:       %s\n", branch)
	fmt.Printf("Commit:       %s  %s\n", short, subject)
	fmt.Printf("Status:       %s\n", formatStatusCounts(counts, colorEnabled, true))
	fmt.Printf("Master:       %s\n", masterBrief)
}

func printAppendedBrokenBlock(dirLine, msg string, colorEnabled bool) {
	statusVal := "error: " + msg
	if colorEnabled {
		statusVal = colorize(statusVal, ansiRed)
	}
	fmt.Printf("Dir:          %s\n", dirLine)
	fmt.Printf("Status:       %s\n", statusVal)
}

func printStatusBlock(root, repoPath string, colorEnabled bool, showRemote bool, fetchEnabled bool, opts statusBlockPrintOpts) error {
	rel := opts.forceRel
	if rel == "" {
		var err error
		rel, err = filepath.Rel(root, repoPath)
		if err != nil {
			return fmt.Errorf("resolve relative repo path: %w", err)
		}
		if rel == "." {
			rel = "."
		}
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
	if opts.showMaster != nil {
		hasMaster = *opts.showMaster
	}
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
	fmt.Printf("Status:       %s\n", statusLine)
	if hasMaster {
		fmt.Printf("Master:       %s\n", masterBrief)
	} else if showRemote && rel == "." {
		remoteLine, err := formatStatusRemoteLine(repoPath, branch, colorEnabled, fetchEnabled, counts)
		if err != nil {
			return err
		}
		fmt.Println(remoteLine)
	}
	return nil
}

func formatStatusRemoteLine(mainRepoPath, currentBranch string, colorEnabled bool, fetchEnabled bool, counts statusCounts) (string, error) {
	upstream, err := gitUpstreamRef(mainRepoPath)
	if err != nil {
		return "", err
	}
	if upstream == "" {
		return "Remote:       (no upstream)", nil
	}
	if fetchEnabled {
		if err := gitFetchUpstreamQuietNoOptionalLocks(mainRepoPath, upstream); err != nil {
			return "Remote:       error: " + err.Error(), nil
		}
	}
	isClean := counts.added == 0 && counts.changed == 0 && counts.renamed == 0 && counts.deleted == 0
	remoteColor := colorEnabled && isClean
	result, err := git.CompareBranches(mainRepoPath, upstream, currentBranch)
	if err != nil {
		return "Remote:       error: " + err.Error(), nil
	}
	return "Remote:       " + FormatRemoteBrief(result, remoteColor), nil
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

func formatCompareWithRemote(mainRepoPath, currentBranch string, colorEnabled bool, fetchEnabled bool) (string, error) {
	upstream, err := gitUpstreamRef(mainRepoPath)
	if err != nil {
		return "", err
	}
	if upstream == "" {
		return "Remote:       (no upstream)", nil
	}
	if fetchEnabled {
		if err := gitFetchUpstreamQuietNoOptionalLocks(mainRepoPath, upstream); err != nil {
			return "", err
		}
	}
	result, err := git.CompareBranches(mainRepoPath, upstream, currentBranch)
	if err != nil {
		return "", err
	}
	return "Remote:       " + FormatRemoteBrief(result, colorEnabled), nil
}

func gitUpstreamRef(repoPath string) (string, error) {
	cmd := gitCommandWithEnv(repoPath, []string{"GIT_OPTIONAL_LOCKS=0"}, "rev-parse", "--abbrev-ref", "@{upstream}")
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
	cmd := gitCommandDir(repoPath, args...)
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
	cmd := gitCommandWithEnv(repoPath, []string{"GIT_OPTIONAL_LOCKS=0"}, args...)
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
	cmd := gitCommand("-C", repoPath, "fetch", "--quiet")
	if out, err := cmd.CombinedOutput(); err != nil {
		if len(out) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git fetch: %w", err)
	}
	return nil
}

func gitFetchQuietNoOptionalLocks(repoPath string) error {
	cmd := gitCommandWithEnv(repoPath, []string{"GIT_OPTIONAL_LOCKS=0"}, "fetch", "--quiet")
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
	cmd := gitCommandWithEnv(repoPath, []string{"GIT_OPTIONAL_LOCKS=0"}, "fetch", "--quiet", remote, branch)
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
