package wrkcli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
	lessflags "github.com/xhd2015/less-flags"
)

// Run executes wrk logic with effective cwd set to cwd.
func Run(cwd string, args []string) error {
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	origWd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	if err := os.Chdir(absCwd); err != nil {
		return fmt.Errorf("chdir: %w", err)
	}
	defer func() { _ = os.Chdir(origWd) }()

	return run(origWd, args)
}

func run(origWd string, args []string) error {
	var done bool
	var list bool
	var confirmFromStdin bool
	var depPath string
	remaining, err := lessflags.Bool("--done", &done).
		Bool("--list", &list).
		Bool("--confirm-from-stdin", &confirmFromStdin).
		String("--dep", &depPath).
		Help("-h, --help", usage()).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			// Help text already printed by Parse; exit 0.
			return nil
		}
		return err
	}

	// remaining holds 0 or 1 positional: the optional <target-dir> (the first
	// positional <dir> is consumed by cmd/wrk extractDir before Run is called).
	// More than one extra positional is an error.
	var targetDir string
	if len(remaining) > 1 {
		return fmt.Errorf("wrk: unexpected arguments")
	}
	if len(remaining) == 1 {
		targetDir = remaining[0]
	}

	if list && done {
		return fmt.Errorf("wrk: --list and --done are mutually exclusive")
	}
	if confirmFromStdin && !done {
		return fmt.Errorf("wrk: --confirm-from-stdin is only valid with --done")
	}
	if depPath != "" && (done || list) {
		return fmt.Errorf("wrk: --dep is mutually exclusive with --done and --list")
	}

	// <target-dir> only applies to the create path. Reject it for any other mode.
	if targetDir != "" && (depPath != "" || list || done) {
		return fmt.Errorf("wrk: unexpected arguments")
	}

	if depPath != "" {
		return runDep(depPath)
	}
	if list {
		return runList()
	}
	if done {
		return runDone(confirmFromStdin)
	}
	return runCreate(origWd, targetDir)
}

// usage returns the wrk help text printed by lessflags when -h/--help is given.
func usage() string {
	return `wrk — git worktree helper

Usage:
  wrk [dir] [target-dir] [flags]

Creates a git worktree from the current directory (or <dir>) and prints its
path. With <target-dir>, the worktree is spawned there instead of the default
location (~/.wrk/worktrees/).

Positional arguments:
  <dir>          optional source checkout to create the worktree from
                 (default: current directory)
  <target-dir>   optional spawn location for the worktree:
                   - missing, parent exists   -> spawn exactly at <target-dir>
                   - existing directory        -> spawn a default-named sub-dir
                   - missing parent            -> error

Flags:
  --done [--confirm-from-stdin]   merge worktree branch back and remove it
  --list                          list worktrees (git worktree list)
  --dep <path>                    spawn a dependency worktree under ./external
  --help, -h                      show this help and exit

Environment:
  WRK_HOME   worktree storage root (default: ~/.wrk)
  WRK_DATE   override the run date (YYYY-MM-DD) used in worktree/branch names
`
}

func runList() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	cmd := exec.Command("git", "-C", cwd, "worktree", "list")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return err
	}
	fmt.Print(string(out))
	return nil
}

func runDone(confirmFromStdin bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	cwd, err = filepath.Abs(cwd)
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

	consumerTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	if err := cascadeExternalWorktrees(consumerTop, confirmFromStdin); err != nil {
		return err
	}

	consumerModDir, err := findGoModDir(cwd, consumerTop)
	if err != nil {
		return err
	}
	modInfo, err := resolve.GetModuleInfo(consumerModDir)
	if err != nil {
		return fmt.Errorf("read consumer go.mod: %w", err)
	}
	if resolve.HasLocalFilesystemReplace(modInfo) {
		return fmt.Errorf("consumer go.mod has a local filesystem replace; resolve replace directives manually before running wrk --done")
	}

	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: checkoutRoot,
		TargetPath: "",
		Remove:     true,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin)
		},
	})
	if err != nil {
		return err
	}
	fmt.Println(result.Message)
	return nil
}

func cascadeExternalWorktrees(consumerTop string, confirmFromStdin bool) error {
	externalDir := filepath.Join(consumerTop, "external")
	entries, err := os.ReadDir(externalDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read external dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		externalPath := filepath.Join(externalDir, entry.Name())
		if !worktree.IsLinked(externalPath) {
			continue
		}
		if err := mergeBackExternalWorktree(externalPath, confirmFromStdin); err != nil {
			return err
		}
	}
	return nil
}

func mergeBackExternalWorktree(externalPath string, confirmFromStdin bool) error {
	_, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: externalPath,
		TargetPath: "",
		Remove:     true,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin)
		},
	})
	if err == nil {
		return nil
	}
	if !errors.Is(err, worktree.ErrConfirmationRequired) {
		return err
	}
	return forceRemoveWorktree(externalPath)
}

func forceRemoveWorktree(wtPath string) error {
	mainRepo, err := worktree.ReadMainRepo(wtPath)
	if err != nil {
		return err
	}
	branch, err := worktree.ReadBranch(wtPath)
	if err != nil {
		return err
	}

	removeCmd := exec.Command("git", "-C", mainRepo, "worktree", "remove", "--force", wtPath)
	if out, err := removeCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove: %w\n%s", err, out)
	}
	if branch != "" && branch != "HEAD" {
		branchCmd := exec.Command("git", "-C", mainRepo, "branch", "-D", branch)
		if out, err := branchCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git branch -D: %w\n%s", err, out)
		}
	}
	return nil
}

func runDep(depArg string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	consumerTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	consumerModDir, err := findGoModDir(cwd, consumerTop)
	if err != nil {
		return err
	}

	depPath, err := filepath.Abs(depArg)
	if err != nil {
		return fmt.Errorf("resolve dep path: %w", err)
	}
	if !worktree.IsInsideWorkTree(depPath) {
		return fmt.Errorf("%s is not a git repository", depPath)
	}

	depSource, err := worktree.ShowToplevel(depPath)
	if err != nil {
		return err
	}
	depMain, err := worktree.ResolveMainRepo(depSource)
	if err != nil {
		return err
	}

	// Resolve the dep module the consumer requires. The dep repo may have no
	// go.mod at its root (e.g. dot-pkgs, whose module lives in go-pkgs/), so
	// scan the whole repo and match a discovered module against the consumer's
	// require/replace paths. depModDir is the module dir relative to the dep
	// repo root ("." for a root go.mod).
	depModDir, _, err := resolveDepModule(consumerModDir, depPath)
	if err != nil {
		return err
	}

	externalDir := filepath.Join(consumerTop, "external")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		return fmt.Errorf("create external dir: %w", err)
	}
	if err := ensureGitignoreExternal(consumerTop); err != nil {
		return err
	}

	consumerMain, err := worktree.ResolveMainRepo(consumerTop)
	if err != nil {
		return err
	}

	baseBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return err
	}
	basename := filepath.Base(depMain)
	branchBase, pathToken, err := resolveNamingInputs(depPath, baseBranch)
	if err != nil {
		return err
	}
	date := resolveWrkDate()

	var externalPath string
	for suffix := 0; suffix < 100; suffix++ {
		candidatePath, branch := externalCandidateNames(consumerTop, basename, branchBase, pathToken, date, suffix)
		if externalCandidateBlocked(consumerMain, candidatePath, branch) {
			continue
		}
		if err := createExternalWorktree(consumerMain, depMain, depPath, candidatePath, branch); err != nil {
			return err
		}
		externalPath = candidatePath
		break
	}
	if externalPath == "" {
		return fmt.Errorf("could not find available external worktree name after 99 attempts")
	}

	// The replace must target the directory holding the dep module's go.mod:
	// the repo root when depModDir is ".", or the sub-module subdir otherwise.
	replaceDir := externalPath
	if depModDir != "." {
		replaceDir = filepath.Join(externalPath, depModDir)
	}
	if _, _, err := replace.ReplaceIn(consumerModDir, replaceDir); err != nil {
		return err
	}
	if err := commands.GoModTidy(&commands.GoModEditOptions{Dir: consumerModDir, Stderr: false, Stdout: false}); err != nil {
		return err
	}

	absPath, err := filepath.Abs(externalPath)
	if err != nil {
		return fmt.Errorf("resolve external worktree path: %w", err)
	}
	fmt.Println(absPath)
	return nil
}

// resolveDepModule scans the dep repo at depPath for Go modules and returns the
// directory (relative to depPath, "." for a root go.mod) and module path of the
// module the consumer requires. This handles dependency repos whose module
// lives in a subdirectory rather than at the repo root.
//
// It returns:
//   - "not a go module" when the dep repo contains no go.mod at all,
//   - "<depPath> is not a dependency of the consumer module" when none of the
//     discovered modules matches a consumer require/replace path.
func resolveDepModule(consumerModDir, depPath string) (modDir string, modPath string, err error) {
	consumerMod, err := resolve.GetModuleInfo(consumerModDir)
	if err != nil {
		return "", "", fmt.Errorf("read consumer go.mod: %w", err)
	}
	// Module paths the consumer depends on, either via require or replace.
	wanted := make(map[string]struct{})
	for _, req := range consumerMod.Require {
		wanted[req.Path] = struct{}{}
	}
	for _, repl := range consumerMod.Replace {
		wanted[repl.Old.Path] = struct{}{}
	}

	modules, err := scan.Scan(depPath, scan.Options{})
	if err != nil {
		return "", "", fmt.Errorf("scan dep modules: %w", err)
	}
	if len(modules) == 0 {
		return "", "", fmt.Errorf("not a go module: %s", depPath)
	}
	for _, m := range modules {
		if m.Path == "" {
			continue
		}
		if _, ok := wanted[m.Path]; ok {
			return m.Dir, m.Path, nil
		}
	}
	return "", "", fmt.Errorf("%s is not a dependency of the consumer module", depPath)
}

func createExternalWorktree(consumerMain, depMain, depPath, externalPath, branch string) error {
	depBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return err
	}
	if depBranch == "HEAD" {
		return fmt.Errorf("dep repository is on a detached HEAD")
	}

	remoteName := "wrk-dep-" + filepath.Base(depMain)
	if err := gitIn(consumerMain, "remote", "add", remoteName, depMain); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}
	defer gitIn(consumerMain, "remote", "remove", remoteName)

	if err := gitIn(consumerMain, "fetch", remoteName, depBranch); err != nil {
		return fmt.Errorf("git fetch dep: %w", err)
	}

	ref := remoteName + "/" + depBranch
	if !branchExists(consumerMain, branch) {
		cmd := exec.Command("git", "-C", consumerMain, "worktree", "add", "-b", branch, externalPath, ref)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git worktree add: %w\n%s", err, out)
		}
		return nil
	}

	cmd := exec.Command("git", "-C", consumerMain, "worktree", "add", "--no-checkout", externalPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w\n%s", err, out)
	}
	checkout := exec.Command("git", "-C", externalPath, "checkout", "--ignore-other-worktrees", branch)
	if out, err := checkout.CombinedOutput(); err != nil {
		_ = exec.Command("git", "-C", consumerMain, "worktree", "remove", "--force", externalPath).Run()
		return fmt.Errorf("git checkout: %w\n%s", err, out)
	}
	return nil
}

func ensureGitignoreExternal(top string) error {
	path := filepath.Join(top, ".gitignore")
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(line) == "/external" {
			return nil
		}
	}
	content := string(data)
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "/external\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	return nil
}

func findGoModDir(cwd, top string) (string, error) {
	dir := cwd
	top = filepath.Clean(top)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if filepath.Clean(dir) == top {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no go.mod found within %s", top)
}

func externalCandidateNames(consumerTop, basename, branchBase, pathToken, date string, suffix int) (path, branch string) {
	name := fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
	branch = branchBase + "-" + date
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
		branch = fmt.Sprintf("%s-%d", branch, suffix)
	}
	return filepath.Join(consumerTop, "external", name), branch
}

func externalCandidateBlocked(mainRepo, wtPath, branch string) bool {
	if _, err := os.Stat(wtPath); err == nil {
		return true
	}
	return branchExists(mainRepo, branch)
}

func gitIn(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, out)
	}
	return nil
}

func runCreate(origWd string, targetDir string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	cwd, err = filepath.Abs(cwd)
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

	mainRepo, err := worktree.ResolveMainRepo(checkoutRoot)
	if err != nil {
		return err
	}

	baseBranch, err := worktree.ReadBranch(cwd)
	if err != nil {
		return err
	}

	date := resolveWrkDate()
	branchBase, pathToken, err := resolveNamingInputs(cwd, baseBranch)
	if err != nil {
		return err
	}
	basename := filepath.Base(mainRepo)

	if targetDir != "" {
		return runCreateTargetDir(origWd, targetDir, checkoutRoot, mainRepo, basename, branchBase, pathToken, date)
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}

	worktreesDir := filepath.Join(wrkHome, "worktrees")
	if err := os.MkdirAll(worktreesDir, 0o755); err != nil {
		return fmt.Errorf("create worktrees dir: %w", err)
	}

	for suffix := 0; suffix < 100; suffix++ {
		wtPath, branch := candidateNames(worktreesDir, basename, branchBase, pathToken, date, suffix)
		if candidateBlocked(mainRepo, wtPath, branch) {
			continue
		}

		if err := createWorktree(checkoutRoot, wtPath, branch, branchExists(mainRepo, branch)); err != nil {
			return err
		}

		absPath, err := filepath.Abs(wtPath)
		if err != nil {
			return fmt.Errorf("resolve worktree path: %w", err)
		}
		fmt.Println(absPath)
		return nil
	}
	return fmt.Errorf("could not find available worktree name after 99 attempts")
}

// runCreateTargetDir handles wrk <dir> <target-dir>. A relative <target-dir> is
// resolved against origWd (the process/shell cwd), NOT the repo dir that Run
// chdir'd into.
func runCreateTargetDir(origWd, targetDir, checkoutRoot, mainRepo, basename, branchBase, pathToken, date string) error {
	// Resolve <target-dir> against the shell cwd (origWd), not the repo dir.
	absTarget := targetDir
	if !filepath.IsAbs(absTarget) {
		absTarget = filepath.Join(origWd, absTarget)
	}
	absTarget = filepath.Clean(absTarget)

	info, err := os.Stat(absTarget)
	if err == nil {
		// Case 2 / file edge: <target-dir> exists.
		if !info.IsDir() {
			return fmt.Errorf("wrk: %s is not a directory", absTarget)
		}
		// Case 2: spawn a default-named sub-dir under <target-dir>, with the
		// usual -N collision handling on both path and branch.
		for suffix := 0; suffix < 100; suffix++ {
			wtPath, branch := candidateNames(absTarget, basename, branchBase, pathToken, date, suffix)
			if candidateBlocked(mainRepo, wtPath, branch) {
				continue
			}
			if err := createWorktree(checkoutRoot, wtPath, branch, branchExists(mainRepo, branch)); err != nil {
				return err
			}
			absPath, err := filepath.Abs(wtPath)
			if err != nil {
				return fmt.Errorf("resolve worktree path: %w", err)
			}
			fmt.Println(absPath)
			return nil
		}
		return fmt.Errorf("could not find available worktree name after 99 attempts")
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("stat target dir: %w", err)
	}

	// <target-dir> does not exist. Case 1 (parent exists) vs case 3 (parent missing).
	parentDir := filepath.Dir(absTarget)
	if _, perr := os.Stat(parentDir); perr != nil {
		if os.IsNotExist(perr) {
			return fmt.Errorf("wrk: %s does not exist", parentDir)
		}
		return fmt.Errorf("stat target parent: %w", perr)
	}

	// Case 1: spawn the worktree exactly at <target-dir> (fixed path, no naming
	// suffix on the path). Branch follows the default convention; if that branch
	// ref already exists, reuse it via the branchPreExists checkout path.
	wtPath := absTarget
	branch := branchBase + "-" + date
	if err := createWorktree(checkoutRoot, wtPath, branch, branchExists(mainRepo, branch)); err != nil {
		return err
	}
	absPath, err := filepath.Abs(wtPath)
	if err != nil {
		return fmt.Errorf("resolve worktree path: %w", err)
	}
	fmt.Println(absPath)
	return nil
}

func resolveWrkHome() (string, error) {
	if v := os.Getenv("WRK_HOME"); v != "" {
		return filepath.Abs(pathfmt.Expand(v))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".wrk"), nil
}

func resolveWrkDate() string {
	if v := os.Getenv("WRK_DATE"); v != "" {
		return v
	}
	return time.Now().Format("2006-01-02")
}

func resolveNamingInputs(cwd, baseBranch string) (branchBase, pathToken string, err error) {
	if baseBranch == "HEAD" {
		hash, err := shortHEAD(cwd)
		if err != nil {
			return "", "", err
		}
		return hash, hash, nil
	}
	return baseBranch, sanitizeBranchToken(baseBranch), nil
}

func shortHEAD(repo string) (string, error) {
	cmd := exec.Command("git", "-C", repo, "rev-parse", "--short=7", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --short=7 HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func sanitizeBranchToken(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func candidateNames(worktreesDir, basename, branchBase, pathToken, date string, suffix int) (path, branch string) {
	name := fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
	branch = branchBase + "-" + date
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
		branch = fmt.Sprintf("%s-%d", branch, suffix)
	}
	return filepath.Join(worktreesDir, name), branch
}

func candidateBlocked(mainRepo, wtPath, branch string) bool {
	if _, err := os.Stat(wtPath); err == nil {
		return true
	}
	return branchExists(mainRepo, branch)
}

func branchExists(repo, branch string) bool {
	cmd := exec.Command("git", "-C", repo, "rev-parse", "--verify", "refs/heads/"+branch)
	return cmd.Run() == nil
}

func createWorktree(sourceDir, wtPath, branch string, branchPreExists bool) error {
	if !branchPreExists {
		cmd := exec.Command("git", "-C", sourceDir, "worktree", "add", "-b", branch, wtPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git worktree add: %w\n%s", err, out)
		}
		return nil
	}

	cmd := exec.Command("git", "-C", sourceDir, "worktree", "add", "--no-checkout", wtPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w\n%s", err, out)
	}

	checkout := exec.Command("git", "-C", wtPath, "checkout", "--ignore-other-worktrees", branch)
	if out, err := checkout.CombinedOutput(); err != nil {
		_ = exec.Command("git", "-C", sourceDir, "worktree", "remove", "--force", wtPath).Run()
		return fmt.Errorf("git checkout: %w\n%s", err, out)
	}
	return nil
}