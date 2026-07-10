package wrkcli

import (
	"bufio"
	"context"
	"errors"
	"fmt"

	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/commands"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/replace"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/resolve"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
	"github.com/xhd2015/dot-pkgs/go-pkgs/wrkcli/storage"
	lessflags "github.com/xhd2015/less-flags"
	"golang.org/x/term"
)

// Run executes wrk logic with args. The first positional argument,
// if present, is the source directory for all modes.
func Run(args []string) error {
	origWd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd: %w", err)
	}
	ctx := newInvocationContext(origWd, args)
	var runErr error
	defer func() {
		exitCode := 0
		if runErr != nil {
			var ece ExitCodeError
			if errors.As(runErr, &ece) {
				exitCode = ece.Code
			} else {
				exitCode = 1
			}
		}
		ctx.finish(exitCode)
	}()
	runErr = run(origWd, args, ctx)
	return runErr
}

func validateWhereFlagArg(args []string) error {
	for i, arg := range args {
		if arg != "--where" {
			continue
		}
		if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
			return fmt.Errorf("wrk: --where requires a path argument")
		}
	}
	return nil
}

func run(origWd string, args []string, ctx *invocationContext) error {
	if len(args) > 0 && args[0] == "skill" {
		wrkHome, err := resolveWrkHome()
		if err != nil {
			return err
		}
		ctx.wrkHome = wrkHome
		ctx.workDir = origWd
		ctx.command = "skill"
		ctx.eventArgs = args[1:]
		if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
			return err
		}
		if err := ctx.autoRecord(); err != nil {
			return err
		}
		return runSkill(origWd, args[1:], wrkHome)
	}
	if hasArg(args, "--bash-integration") {
		ctx.skipEvent = true
		return runBashIntegration(args)
	}
	if hasArg(args, "--interceptor") {
		return runInterceptorMgmt(origWd, args, ctx)
	}

	if err := validateWhereFlagArg(args); err != nil {
		return err
	}

	var done bool
	var mergeBack bool
	var list bool
	var status bool
	var repos bool
	var projects bool
	var colorFlag bool
	var fetchFlag bool
	var verbose bool
	var addPath *string
	var removePath *string
	var confirmFromStdin bool
	var assumeYes bool
	var noInModuleReplace bool
	var depPath string
	var allDeps bool
	var dryRun bool
	var taskDesc *string
	var setTaskDesc *string
	var wherePath *string
	var noCd bool
	var cd bool
	var noInterceptor bool
	var mainFlag bool
	// *string targets: nil = flag absent; non-nil empty = present but empty.
	remaining, err := lessflags.Bool("--done", &done).
		Bool("--merge-back", &mergeBack).
		Bool("-l,--list", &list).
		Bool("--status", &status).
		Bool("--repos", &repos).
		Bool("--projects", &projects).
		Bool("--fetch", &fetchFlag).
		Bool("-v,--verbose", &verbose).
		Bool("--color", &colorFlag).
		String("--add", &addPath).
		String("--rm", &removePath).
		Bool("--confirm-from-stdin", &confirmFromStdin).
		Bool("-y,--yes", &assumeYes).
		Bool("--no-in-module-replace", &noInModuleReplace).
		Bool("--no-cd", &noCd).
		Bool("--cd", &cd).
		Bool("--no-interceptor", &noInterceptor).
		Bool("--main", &mainFlag).
		Bool("--all-deps", &allDeps).
		Bool("--dry-run", &dryRun).
		String("--dep", &depPath).
		String("-t,--task", &taskDesc).
		String("--set-task", &setTaskDesc).
		String("--where", &wherePath).
		Help("-h,--help", usage()).
		HelpNoExit().
		Parse(args)
	if err != nil {
		if errors.Is(err, lessflags.ErrHelp) {
			// Help text already printed by Parse; exit 0.
			ctx.skipEvent = true
			return nil
		}
		return err
	}

	taskFlagSet := taskDesc != nil
	setTaskFlagSet := setTaskDesc != nil
	addFlagSet := addPath != nil
	removeFlagSet := removePath != nil
	whereFlagSet := wherePath != nil

	ctx.command = resolveCommand(projects, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, done, list, status, repos, mergeBack, depPath, allDeps, cd, mainFlag)
	ctx.eventArgs = extractEventArgs(args, remaining)

	setInvocationVerbose(verbose)
	worktree.GitVerboseLogger = logGitCommand
	defer func() {
		setInvocationVerbose(false)
		worktree.GitVerboseLogger = nil
	}()

	if fetchFlag && !projects && !status {
		return fmt.Errorf("wrk: --fetch is only valid with --projects or --status")
	}

	// remaining holds 0, 1, or 2 positionals:
	//   remaining[0] = sourceDir (valid for ALL modes — cwd when absent)
	//   remaining[1] = spawnTarget (create-only, was targetDir)
	// More than two positionals is an error.
	if len(remaining) > 2 {
		return fmt.Errorf("wrk: unexpected arguments")
	}
	var sourceDir string
	var spawnTarget string
	if len(remaining) >= 1 {
		sourceDir = remaining[0]
	}
	if len(remaining) == 2 {
		spawnTarget = remaining[1]
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}
	ctx.wrkHome = wrkHome

	// --cd requires exactly one path positional before defaulting workDir to cwd.
	if cd {
		if len(remaining) == 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: --cd requires a path argument")
		}
		if len(remaining) > 1 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}

	// --main takes no path positional. Mutual exclusion is checked later; if another
	// mode flag is also set, prefer that error over unexpected arguments.
	if mainFlag {
		otherMode := done || list || status || repos || projects || addFlagSet || removeFlagSet ||
			whereFlagSet || depPath != "" || allDeps || dryRun || mergeBack || taskFlagSet ||
			setTaskFlagSet || fetchFlag || noCd || cd || spawnTarget != ""
		if otherMode {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: --main is mutually exclusive with other modes")
		}
		if len(remaining) > 0 {
			ctx.workDir = origWd
			if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
				return err
			}
			return fmt.Errorf("wrk: unexpected arguments")
		}
	}

	// Resolve sourceDir to absolute; default to process cwd when absent.
	// Passed to every sub-command as workDir instead of using os.Getwd/Chdir.
	createMode := isCreateMode(projects, addFlagSet, removeFlagSet, setTaskFlagSet, whereFlagSet, repos, status, depPath, allDeps, list, done, mergeBack, cd, mainFlag)
	dirHint := &DirHintOptions{
		RawArgs:     args,
		Positionals: remaining,
	}
	// Basename fallback: create/status/list/repos/--cd. --main uses cwd only.
	workDir, err := resolveSourceWorkDir(origWd, sourceDir, createMode || status || list || repos || cd, wrkHome, dirHint)
	if err != nil {
		return err
	}
	ctx.workDir = workDir
	if err := storage.ResetEventsIfDoctest(wrkHome); err != nil {
		return err
	}
	if err := ctx.autoRecord(); err != nil {
		return err
	}

	if addFlagSet && strings.TrimSpace(*addPath) == "" {
		return fmt.Errorf("wrk: --add requires a path argument")
	}
	if removeFlagSet && strings.TrimSpace(*removePath) == "" {
		return fmt.Errorf("wrk: --rm requires a path argument")
	}
	if whereFlagSet && strings.TrimSpace(*wherePath) == "" {
		return fmt.Errorf("wrk: --where requires a path argument")
	}

	// --set-task is mutually exclusive with all other modes.
	if setTaskFlagSet && strings.TrimSpace(*setTaskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	// --set-task is mutually exclusive with all other modes.
	if setTaskFlagSet && (taskFlagSet || done || list || status || repos || projects || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --set-task is mutually exclusive with other flags")
	}
	if setTaskFlagSet {
		return runSetTask(workDir, *setTaskDesc, assumeYes, noCd)
	}

	if taskFlagSet && strings.TrimSpace(*taskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	// --task is only valid with create mode.
	if taskFlagSet && (done || list || status || repos || projects || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || mergeBack || cd || mainFlag) {
		return fmt.Errorf("wrk: --task is mutually exclusive with --done, --merge-back, --list, --status, --repos, --projects, --add, --rm, --where, --dep and --all-deps")
	}

	if list && done {
		return fmt.Errorf("wrk: --list and --done are mutually exclusive")
	}
	if list && mergeBack {
		return fmt.Errorf("wrk: --list and --merge-back are mutually exclusive")
	}
	if done && mergeBack {
		return fmt.Errorf("wrk: --done and --merge-back are mutually exclusive")
	}
	if repos && (done || list || status || projects || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --repos is mutually exclusive with other modes")
	}
	if projects && (done || list || status || repos || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --projects is mutually exclusive with other modes")
	}
	if addFlagSet && (done || list || status || repos || projects || removeFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --add is mutually exclusive with other modes")
	}
	if removeFlagSet && (done || list || status || repos || projects || addFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --rm is mutually exclusive with other modes")
	}
	if whereFlagSet && (done || list || status || repos || projects || addFlagSet || removeFlagSet || depPath != "" || allDeps || dryRun || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || fetchFlag || cd || mainFlag) {
		return fmt.Errorf("wrk: --where is mutually exclusive with other modes")
	}
	if cd && (done || list || status || repos || projects || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || fetchFlag || noCd || mainFlag) {
		return fmt.Errorf("wrk: --cd is mutually exclusive with other modes")
	}
	if mainFlag && (done || list || status || repos || projects || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || mergeBack || taskFlagSet || setTaskFlagSet || spawnTarget != "" || fetchFlag || noCd || cd) {
		return fmt.Errorf("wrk: --main is mutually exclusive with other modes")
	}
	if status && (done || list || projects || addFlagSet || removeFlagSet || whereFlagSet || depPath != "" || allDeps || dryRun || spawnTarget != "" || cd || mainFlag) {
		return fmt.Errorf("wrk: --status is mutually exclusive with other modes")
	}
	if confirmFromStdin && !done && !mergeBack {
		return fmt.Errorf("wrk: --confirm-from-stdin is only valid with --done or --merge-back")
	}
	if noInModuleReplace && !done {
		return fmt.Errorf("wrk: --no-in-module-replace is only valid with --done")
	}
	if depPath != "" && (done || list || mergeBack || cd || mainFlag) {
		return fmt.Errorf("wrk: --dep is mutually exclusive with --done, --merge-back and --list")
	}
	if allDeps && (depPath != "" || done || list || mergeBack || cd || mainFlag) {
		return fmt.Errorf("wrk: --all-deps is mutually exclusive with --dep, --done, --merge-back and --list")
	}
	if dryRun && !allDeps {
		return fmt.Errorf("wrk: --dry-run is only valid with --all-deps")
	}

	// spawnTarget only applies to the create path. Reject for any other mode.
	if spawnTarget != "" && (depPath != "" || allDeps || list || status || repos || projects || addFlagSet || removeFlagSet || whereFlagSet || done || mergeBack || cd || mainFlag) {
		return fmt.Errorf("wrk: unexpected arguments")
	}
	if whereFlagSet && len(remaining) > 0 {
		return fmt.Errorf("wrk: unexpected arguments")
	}

	if projects {
		colorEnabled := term.IsTerminal(int(os.Stdout.Fd())) || colorFlag
		return runProjects(wrkHome, colorEnabled, fetchFlag)
	}
	if addFlagSet {
		return runAdd(wrkHome, *addPath)
	}
	if removeFlagSet {
		return runRemove(wrkHome, *removePath)
	}
	if whereFlagSet {
		return runWhere(wrkHome, *wherePath)
	}
	if mainFlag {
		return runMain(workDir)
	}
	if cd {
		return runCd(workDir)
	}
	if repos {
		return runRepos(workDir)
	}
	if status {
		colorEnabled := term.IsTerminal(int(os.Stdout.Fd())) || colorFlag
		return runStatus(workDir, colorEnabled, fetchFlag)
	}
	if depPath != "" {
		return runDep(workDir, depPath, wrkHome, args)
	}
	if allDeps {
		return runAllDeps(workDir, dryRun)
	}
	if list {
		return runList(workDir)
	}
	if done {
		return runDone(workDir, confirmFromStdin, assumeYes, noInModuleReplace, noCd)
	}
	if mergeBack {
		return runMergeBack(workDir, confirmFromStdin, assumeYes)
	}
	task := ""
	if taskDesc != nil {
		task = *taskDesc
	}
	// Optional create-mode interceptor: expand argv/vars from config.json and
	// exec instead of native worktree create. Escape: --no-interceptor /
	// WRK_NO_INTERCEPTOR=1. No-op on non-create modes (flag already parsed).
	if !noInterceptor && os.Getenv("WRK_NO_INTERCEPTOR") != "1" {
		ic, err := loadCreateInterceptor(wrkHome)
		if err != nil {
			return err
		}
		if ic != nil && ic.Enabled {
			return runCreateInterceptor(ic, createInterceptorInput{
				wrkHome:     wrkHome,
				workDir:     workDir,
				origWd:      origWd,
				source:      sourceDir,
				spawnTarget: spawnTarget,
				task:        task,
				args:        args,
			})
		}
	}
	return runCreate(workDir, origWd, spawnTarget, task, noCd)
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
  --merge-back [--confirm-from-stdin]  merge worktree branch back WITHOUT removing it
  --done --no-in-module-replace   block --done on ANY local replace (strict)
  --list                          list worktrees (git worktree list)
  --status                        show status for git repos under this checkout
  --repos                         list git repos under this checkout
  --projects                      list recorded main repository paths
  --fetch                         with --projects or --status: fetch upstream before Remote: compare
  -v, --verbose                   log major git commands to stderr
  --add <dir>                     manually record a main repository path
  --rm <dir>                      remove a recorded main repository path
  --where <basename>              look up saved project path(s) by basename
  --cd <path|basename>            jump into directory (in-place follow-up or interactive shell)
  --main                          open nested shell at main repository root for this checkout
  --dep <path>                    spawn a dependency worktree under ./external
  --all-deps                      link every required dep from registered projects
  --dry-run                       with --all-deps: plan only, no writes
  --task <desc>                   append task slug to worktree/branch names
  --set-task <desc>               rename worktree/branch to match new task
  -y, --yes                       auto-confirm Y/n prompts (own worktree; cascade on TTY only)
  --no-cd                         do not write shell follow-up cd lines (for bash auto-cd wrapper)
  --no-interceptor                skip create.interceptor and use native create (no-op on non-create)
  --help, -h                      show this help and exit

Interceptor management:
  wrk --interceptor --status      show absent|disabled|enabled, path, argv0
  wrk --interceptor --show        pretty-print create.interceptor JSON
  wrk --interceptor --path        print absolute path to $WRK_HOME/config.json
  wrk --interceptor --enable      set enabled=true (requires existing block)
  wrk --interceptor --disable     set enabled=false
  wrk --interceptor --init [--force]  write disabled neutral stub
  wrk --interceptor --check       validate interceptor when present
  wrk --interceptor --dry-run [--] [create-args...]  expand argv without exec

Skill commands:
  wrk skill --list|-l             list available skills (wrk)
  wrk skill --show [--header]     print wrk SKILL.md (full or YAML header only)
  wrk skill --install [flags]     install wrk SKILL.md to agent directories

Environment:
  WRK_HOME              worktree storage root (default: ~/.wrk)
  WRK_DATE              override the run date (YYYY-MM-DD) used in worktree/branch names
  WRK_NO_INTERCEPTOR    set to 1 to skip create.interceptor (same as --no-interceptor)
`
}

func runProjects(wrkHome string, colorEnabled bool, fetchEnabled bool) error {
	endPerf := beginProjectsPerfRun()
	defer endPerf()

	paths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return nil
	}

	results := make([]projectStatusData, len(paths))
	done := make([]bool, len(paths))

	nextPrint := 0
	printedAny := false
	var mu sync.Mutex
	flush := func() {
		for nextPrint < len(paths) && done[nextPrint] {
			if printedAny {
				fmt.Println()
			}
			printProjectStatusFromData(results[nextPrint], colorEnabled)
			printedAny = true
			nextPrint++
		}
	}

	workers := minInt(projectsProjectWorkers(), len(paths))
	jobs := make(chan int, len(paths))
	for i := range paths {
		jobs <- i
	}
	close(jobs)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				p := paths[i]
				endProject := beginProjectPerf(p)
				data, _ := gatherProjectStatus(p, colorEnabled, fetchEnabled)
				endProject()

				mu.Lock()
				results[i] = data
				done[i] = true
				flush()
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return nil
}

func runAdd(wrkHome, addDir string) error {
	abs, err := filepath.Abs(addDir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("wrk: %s does not exist", abs)
		}
		return fmt.Errorf("stat dir: %w", err)
	}
	if !worktree.IsInsideWorkTree(abs) {
		return fmt.Errorf("%s is not a git repository", abs)
	}
	top, err := worktree.ShowToplevel(abs)
	if err != nil {
		return err
	}
	mainRepo, err := worktree.ResolveMainRepo(top)
	if err != nil {
		return err
	}
	mainRepo = storage.NormalizePath(mainRepo)
	if err := storage.RecordProject(wrkHome, mainRepo, storage.SourceManual); err != nil {
		return err
	}
	fmt.Println(mainRepo)
	return nil
}

func runRemove(wrkHome, removeDir string) error {
	abs, err := filepath.Abs(removeDir)
	if err != nil {
		return fmt.Errorf("resolve dir: %w", err)
	}
	mainRepoPath := storage.NormalizePath(abs)
	if _, err := os.Stat(abs); err == nil {
		if worktree.IsInsideWorkTree(abs) {
			top, err := worktree.ShowToplevel(abs)
			if err != nil {
				return err
			}
			mainRepo, err := worktree.ResolveMainRepo(top)
			if err != nil {
				return err
			}
			mainRepoPath = storage.NormalizePath(mainRepo)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat dir: %w", err)
	}
	removed, err := storage.RemoveProject(wrkHome, mainRepoPath)
	if err != nil {
		return err
	}
	if removed {
		fmt.Println(mainRepoPath)
	}
	return nil
}

func runList(workDir string) error {
	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsInsideWorkTree(cwd) {
		return fmt.Errorf("%s is not a git repository", cwd)
	}

	cmd := gitCommand("-C", cwd, "worktree", "list")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			return fmt.Errorf("%s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return err
	}
	outStr := string(out)
	if len(outStr) > 0 && !strings.HasSuffix(outStr, "\n") {
		outStr += "\n"
	}
	fmt.Print(outStr)
	return nil
}

func runDone(workDir string, confirmFromStdin, assumeYes, noInModuleReplace, noCd bool) error {
	// Shell process cwd (inherited from interactive shell), not merely workDir.
	// Used after remove to decide whether auto-cd is needed.
	shellCwd, _ := os.Getwd()

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

	consumerTop, err := worktree.ShowToplevel(cwd)
	if err != nil {
		return err
	}
	if err := checkCascadeNonInteractive(consumerTop, checkoutRoot); err != nil {
		return err
	}
	if err := cascadeLinkedWorktrees(consumerTop, checkoutRoot, confirmFromStdin, assumeYes); err != nil {
		return err
	}

	// Guard: classify every local filesystem replace under the checkout (main
	// or sub-module). wrk --dep/--all-deps write replace => ./external/... and
	// --done's cascade removes those external worktrees, so a remaining local
	// replace would dangle — those (extra-repo) block. An intra-repo replace
	// (target exists and shares the consumer's toplevel, e.g. ../../ or ./sub
	// pointing back into the same repo) is stable, so under the default lenient
	// guard it only warns and --done proceeds; --no-in-module-replace makes
	// every local replace block. Scanning every module (not just the nearest
	// go.mod) also catches sub-module replaces a single upward lookup would
	// miss. A checkout with no go.mod yields zero modules → guard is a no-op →
	// MergeBack proceeds (it is pure git).
	if err := blockIfLocalReplace(consumerTop, noInModuleReplace); err != nil {
		return err
	}

	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: checkoutRoot,
		TargetPath: "",
		Remove:     true,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	if err != nil {
		return err
	}
	fmt.Println(result.Message)
	// After successful remove: write follow-up cd only if shell cwd is gone
	// (user was inside the removed worktree). Surviving sibling/main stays put.
	if result.Action != "aborted" && result.Action != "dry-run" {
		if err := writeFollowupCDIfCwdMissing(noCd, shellCwd, result.TargetPath); err != nil {
			return err
		}
	}
	return nil
}

func runMergeBack(workDir string, confirmFromStdin, assumeYes bool) error {
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

	result, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: checkoutRoot,
		TargetPath: "",
		Remove:     false,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	if err != nil {
		return err
	}
	fmt.Println(result.Message)
	return nil
}

func checkCascadeNonInteractive(consumerTop, checkoutRoot string) error {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	repos, err := discoverStatusRepos(context.Background(), consumerTop)
	if err != nil {
		return err
	}
	cleanCheckout := filepath.Clean(checkoutRoot)
	for _, repo := range repos {
		if repo.RepoType == scan_repo.RepoTypeMain {
			continue
		}
		if repo.RepoType != scan_repo.RepoTypeWorktree {
			continue
		}
		if !worktree.IsLinked(repo.Path) {
			continue
		}
		if filepath.Clean(repo.Path) == cleanCheckout {
			continue
		}
		mainRepo, err := worktree.ResolveMainRepo(repo.Path)
		if err != nil {
			return err
		}
		inclusion, err := worktree.HeadIncludedInMain(mainRepo, repo.Path)
		if err != nil {
			return err
		}
		if inclusion.Relation == "ahead" || inclusion.Relation == "diverged" {
			return fmt.Errorf("wrk --done: cannot cascade merge-back non-interactively: linked worktree %s is %s and needs confirmation", repo.Path, inclusion.Relation)
		}
	}
	return nil
}

func cascadeLinkedWorktrees(consumerTop, checkoutRoot string, confirmFromStdin, assumeYes bool) error {
	repos, err := discoverStatusRepos(context.Background(), consumerTop)
	if err != nil {
		return err
	}

	cleanCheckout := filepath.Clean(checkoutRoot)
	for _, repo := range repos {
		if repo.RepoType == scan_repo.RepoTypeMain {
			if filepath.Clean(repo.Path) != filepath.Clean(consumerTop) {
				fmt.Fprintf(os.Stderr, "warning: skipping nested main repo %s\n", repo.Path)
			}
			continue
		}
		if repo.RepoType != scan_repo.RepoTypeWorktree {
			continue
		}
		if !worktree.IsLinked(repo.Path) {
			continue
		}
		if filepath.Clean(repo.Path) == cleanCheckout {
			continue
		}
		if err := mergeBackExternalWorktree(repo.Path, confirmFromStdin, assumeYes); err != nil {
			return err
		}
	}
	return nil
}

// mergeBackExternalWorktree merge-backs (or removes) an external dependency
// worktree during the --done cascade.
//
// External dep worktrees are worktrees of the DEP repo (registered under
// <depMain>/.git/worktrees/, per createExternalWorktree's git -C depMain worktree
// add), so MergeBack resolves the owning main repo from the worktree's .git
// gitdir (the dep main) and merges the dep branch back into the dep repo — the
// branch shares the dep's history, so the merge-base check resolves. This
// ensures dep work committed on the external worktree is merged back into the
// dep repo before the worktree is removed. Relation to dep main: already-included
// → remove only; ahead/diverged → prompt (via confirmFromStdin). A
// non-interactive ahead/diverged worktree errors (no force-removal fallback).
func mergeBackExternalWorktree(externalPath string, confirmFromStdin, assumeYes bool) error {
	_, err := worktree.MergeBack(worktree.MergeBackOptions{
		SourcePath: externalPath,
		TargetPath: "",
		Remove:     true,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return worktree.PromptConfirmPlan(plan, confirmFromStdin, assumeYes)
		},
	})
	return err
}

func runDep(workDir string, depArg string, wrkHome string, rawArgs []string) error {
	cwd, err := filepath.Abs(workDir)
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
	depPath, err := resolveDirArg(depArg, true, wrkHome, &DirHintOptions{
		RawArgs: rawArgs,
		DepMode: true,
	})
	if err != nil {
		return err
	}
	depMain, err := worktree.ResolveMainRepo(depPath)
	if err != nil {
		return err
	}

	// Scan both dep and consumer repos for Go modules. The consumer may have
	// no root go.mod (e.g. dot-pkgs, whose module lives in go-pkgs/), so we
	// scan the whole tree and replace+tidy in every matching consumer module.
	depModules, err := scan.Scan(depPath, scan.Options{})
	if err != nil {
		return fmt.Errorf("scan dep modules: %w", err)
	}
	if len(depModules) == 0 {
		return fmt.Errorf("not a go module: %s", depPath)
	}

	consumerModules, err := scan.Scan(consumerTop, scan.Options{})
	if err != nil {
		return fmt.Errorf("scan consumer modules: %w", err)
	}

	// depModDir is the dep module's directory relative to the dep repo root
	// (e.g. "." for a root go.mod, "go-pkgs" for a sub-module). It's resolved
	// from the first consumer module that depends on the dep.
	type consumerMatch struct{ dir string }
	var matchingConsumerDirs []consumerMatch
	var depModDir string
	for _, cm := range consumerModules {
		wanted := make(map[string]struct{})
		for _, req := range cm.Requires {
			wanted[req.Path] = struct{}{}
		}
		for _, repl := range cm.Replaces {
			wanted[repl.OldPath] = struct{}{}
		}
		for _, dm := range depModules {
			if dm.Path == "" {
				continue
			}
			if _, ok := wanted[dm.Path]; ok {
				matchingConsumerDirs = append(matchingConsumerDirs, consumerMatch{
					dir: filepath.Join(consumerTop, cm.Dir),
				})
				if depModDir == "" {
					depModDir = dm.Dir
				}
			}
		}
	}
	if len(matchingConsumerDirs) == 0 {
		return fmt.Errorf("%s is not a dependency of any consumer module", depPath)
	}

	externalDir := filepath.Join(consumerTop, "external")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		return fmt.Errorf("create external dir: %w", err)
	}
	if err := ensureGitignoreExternal(consumerTop); err != nil {
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
		// The external worktree + its branch are owned by the DEP repo, so the
		// branch-collision check must run against depMain (not the consumer).
		if externalCandidateBlocked(depMain, candidatePath, branch) {
			continue
		}
		if err := createExternalWorktree(depMain, depPath, candidatePath, branch); err != nil {
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
	for _, m := range matchingConsumerDirs {
		if _, _, err := replace.ReplaceIn(m.dir, replaceDir); err != nil {
			return err
		}
		if err := commands.GoModTidy(&commands.GoModEditOptions{Dir: m.dir, Stderr: false, Stdout: false}); err != nil {
			return err
		}
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

// planExternalWorktreePath is the read-only planner for an external dep
// worktree: it resolves the dep's main repo, basename, branch base, path token,
// date and consumer main repo, then runs the suffix loop
// (externalCandidateNames + externalCandidateBlocked) to return the first
// non-blocked candidate external worktree path. It performs NO writes: no
// MkdirAll(external/), no ensureGitignoreExternal, no createExternalWorktree.
// It may call read-only git helpers (ShowToplevel, ResolveMainRepo, ReadBranch,
// resolveNamingInputs) which only run git rev-parse / git symbolic-ref.
func planExternalWorktreePath(consumerTop, depPath string) (externalPath string, err error) {
	depSource, err := worktree.ShowToplevel(depPath)
	if err != nil {
		return "", err
	}
	depMain, err := worktree.ResolveMainRepo(depSource)
	if err != nil {
		return "", err
	}

	baseBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return "", err
	}
	basename := filepath.Base(depMain)
	branchBase, pathToken, err := resolveNamingInputs(depPath, baseBranch)
	if err != nil {
		return "", err
	}
	date := resolveWrkDate()

	for suffix := 0; suffix < 100; suffix++ {
		candidatePath, branch := externalCandidateNames(consumerTop, basename, branchBase, pathToken, date, suffix)
		// Branch-collision check runs against depMain: the external worktree's
		// branch lives in the dep repo (see createExternalWorktree).
		if externalCandidateBlocked(depMain, candidatePath, branch) {
			continue
		}
		return candidatePath, nil
	}
	return "", fmt.Errorf("could not find available external worktree name after 99 attempts")
}

// createExternalWorktreeForRepo materializes the external worktree for the dep
// repo resolved from depPath under {consumerTop}/external/ and returns its path.
// It plans the path via planExternalWorktreePath (so dry-run and real runs
// agree on naming), then creates the external dir, ensures .gitignore, and adds
// the worktree. It does NOT add a replace directive or run tidy. Used by
// runAllDeps (one worktree per repo, with per-module replaces added separately).
func createExternalWorktreeForRepo(consumerTop, depPath string) (externalPath string, err error) {
	externalPath, err = planExternalWorktreePath(consumerTop, depPath)
	if err != nil {
		return "", err
	}

	externalDir := filepath.Join(consumerTop, "external")
	if err := os.MkdirAll(externalDir, 0o755); err != nil {
		return "", fmt.Errorf("create external dir: %w", err)
	}
	if err := ensureGitignoreExternal(consumerTop); err != nil {
		return "", err
	}

	depSource, err := worktree.ShowToplevel(depPath)
	if err != nil {
		return "", err
	}
	depMain, err := worktree.ResolveMainRepo(depSource)
	if err != nil {
		return "", err
	}

	baseBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return "", err
	}
	basename := filepath.Base(depMain)
	branchBase, pathToken, err := resolveNamingInputs(depPath, baseBranch)
	if err != nil {
		return "", err
	}
	date := resolveWrkDate()

	for suffix := 0; suffix < 100; suffix++ {
		candidatePath, branch := externalCandidateNames(consumerTop, basename, branchBase, pathToken, date, suffix)
		if candidatePath != externalPath {
			// planExternalWorktreePath already selected the first non-blocked
			// candidate; later suffixes are never needed here.
			continue
		}
		if err := createExternalWorktree(depMain, depPath, candidatePath, branch); err != nil {
			return "", err
		}
		break
	}
	return externalPath, nil
}

func runAllDeps(workDir string, dryRun bool) error {
	cwd, err := filepath.Abs(workDir)
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

	// Scan the consumer tree for all Go modules (supports repos like dot-pkgs
	// whose module lives in a subdirectory with no root go.mod).
	consumerModules, err := scan.Scan(consumerTop, scan.Options{})
	if err != nil {
		return fmt.Errorf("scan consumer modules: %w", err)
	}

	type consumerModInfo struct {
		dir             string // abs path to the module's go.mod directory
		modulePath      string
		required        map[string]bool // dep paths this module requires
		alreadyReplaced map[string]bool // dep paths already replaced in this module
	}
	var consumerMods []consumerModInfo
	allRequired := make(map[string]bool)
	allAlreadyReplaced := make(map[string]bool)
	allConsumerModules := make(map[string]bool)

	for _, cm := range consumerModules {
		dir := filepath.Join(consumerTop, cm.Dir)
		info := consumerModInfo{
			dir:             dir,
			modulePath:      cm.Path,
			required:        make(map[string]bool),
			alreadyReplaced: make(map[string]bool),
		}
		if cm.Path != "" {
			allConsumerModules[cm.Path] = true
		}
		for _, req := range cm.Requires {
			info.required[req.Path] = true
			allRequired[req.Path] = true
		}
		for _, repl := range cm.Replaces {
			info.alreadyReplaced[repl.OldPath] = true
			allAlreadyReplaced[repl.OldPath] = true
		}
		consumerMods = append(consumerMods, info)
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}

	projectPaths, err := storage.ListProjects(wrkHome)
	if err != nil {
		return err
	}

	type linkedDep struct {
		modulePath   string
		externalPath string // replace target (repo-root external path, or a sub-dir for nested sub-modules)
	}
	seen := make(map[string]bool)
	var linked []linkedDep
	tidied := make(map[string]bool)
	for _, projectPath := range projectPaths {
		if _, err := os.Stat(projectPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}
		if !worktree.IsMainRepo(projectPath) {
			continue
		}
		// mod/scan finds all modules in the repo (root + nested sub-modules) in
		// process, with vendor/testdata/gitignore skips. On error (e.g. unreadable
		// go.mod) skip the repo.
		modules, err := scan.Scan(projectPath, scan.Options{})
		if err != nil {
			continue
		}
		// Collect matched modules first so the repo's worktree is only created
		// when at least one module matches (and shared across all of them).
		var matched []scan.Module
		for _, m := range modules {
			if m.Path == "" || allConsumerModules[m.Path] {
				continue
			}
			if !allRequired[m.Path] || allAlreadyReplaced[m.Path] || seen[m.Path] {
				continue
			}
			matched = append(matched, m)
		}
		if len(matched) == 0 {
			continue
		}

		if dryRun {
			// Dry-run: compute the planned external path (read-only) and per-
			// module replace targets, but write nothing (no
			// createExternalWorktree, no GoModEditReplace, no tidy, no
			// gitignore).
			externalPath, err := planExternalWorktreePath(consumerTop, projectPath)
			if err != nil {
				return err
			}
			for _, m := range matched {
				target := externalPath
				if m.Dir != "." {
					target = filepath.Join(externalPath, filepath.FromSlash(m.Dir))
				}
				seen[m.Path] = true
				linked = append(linked, linkedDep{modulePath: m.Path, externalPath: target})
			}
			continue
		}
		// Real run: materialize the planned external worktree.
		externalPath, err := createExternalWorktreeForRepo(consumerTop, projectPath)
		if err != nil {
			return err
		}
		for _, m := range matched {
			// m.Dir is "." for the repo root module, or a slash-joined sub-dir
			// (e.g. "services/dep") for a nested sub-module. The replace target
			// is the sub-module's directory inside the external worktree.
			target := externalPath
			if m.Dir != "." {
				target = filepath.Join(externalPath, filepath.FromSlash(m.Dir))
			}
			// Replace in every consumer module that requires this dep.
			for _, cm := range consumerMods {
				if !cm.required[m.Path] || cm.alreadyReplaced[m.Path] {
					continue
				}
				opts := &commands.GoModEditOptions{Dir: cm.dir, Stderr: false, Stdout: false}
				if err := commands.GoModEditReplace(m.Path, target, opts); err != nil {
					return err
				}
				tidied[cm.dir] = true
			}
			seen[m.Path] = true
			linked = append(linked, linkedDep{modulePath: m.Path, externalPath: target})
		}
	}

	if !dryRun {
		for dir := range tidied {
			if err := commands.GoModTidy(&commands.GoModEditOptions{Dir: dir, Stderr: false, Stdout: false}); err != nil {
				return err
			}
		}
	}

	prefix := ""
	if dryRun {
		prefix = "would: "
	}
	for _, d := range linked {
		rel, err := filepath.Rel(consumerTop, d.externalPath)
		if err != nil {
			return fmt.Errorf("rel external path: %w", err)
		}
		fmt.Printf("%swrk %s at ./%s\n", prefix, d.modulePath, rel)
	}
	fmt.Printf("%swrked %d deps\n", prefix, len(linked))
	return nil
}

// createExternalWorktree spawns the external dep worktree as a worktree of the
// DEP repo (not the consumer). The dep already holds its own objects, so no
// remote/fetch into the consumer is needed; the worktree and its branch are
// registered under <depMain>/.git/worktrees/ — where they semantically belong.
// This also lets `wrk --done` cascade merge dep changes back into the dep repo:
// the dep branch shares the dep's history, so merge-base resolves (the previous
// consumer-owned design failed with "failed to find merge base" because the dep
// branch and the consumer's main had no common ancestor).
func createExternalWorktree(depMain, depPath, externalPath, branch string) error {
	depBranch, err := worktree.ReadBranch(depPath)
	if err != nil {
		return err
	}
	if depBranch == "HEAD" {
		return fmt.Errorf("dep repository is on a detached HEAD")
	}

	if !branchExists(depMain, branch) {
		cmd := gitCommand("-C", depMain, "worktree", "add", "-b", branch, externalPath, depBranch)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git worktree add: %w\n%s", err, out)
		}
		return nil
	}

	// Branch already exists in the dep repo (e.g. an earlier --dep created it):
	// add a linked worktree on that branch without checkout, then check it out.
	cmd := gitCommand("-C", depMain, "worktree", "add", "--no-checkout", externalPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree add: %w\n%s", err, out)
	}
	checkout := gitCommand("-C", externalPath, "checkout", "--ignore-other-worktrees", branch)
	if out, err := checkout.CombinedOutput(); err != nil {
		_ = gitCommand("-C", depMain, "worktree", "remove", "--force", externalPath).Run()
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

// blockIfLocalReplace scans every Go module under top (main + sub-modules) and
// classifies each filesystem/local replace directive. A checkout with no go.mod
// yields zero modules and is allowed: no go.mod means no replace can exist, and
// MergeBack itself is pure git.
//
// A replace is intra-repo when its target resolves to an existing directory
// that shares the consumer's git toplevel (a ../../ or ./sub reference back
// into the same repo); otherwise it is extra-repo (./external dep worktree,
// non-existent target, absolute or sibling-repo path).
//
// Under the default lenient guard, intra-repo replaces only warn (printed to
// stderr) and --done proceeds; extra-repo replaces block. When noInModuleReplace
// is set, every local replace blocks (fully strict).
func blockIfLocalReplace(top string, noInModuleReplace bool) error {
	issues, err := replace.CheckLocalReplaces(top)
	if err != nil {
		return fmt.Errorf("check local replaces under %s: %w", top, err)
	}

	for _, issue := range issues {
		hasExtra := !issue.IsIntraRepo

		if hasExtra || noInModuleReplace {
			var b strings.Builder
			b.WriteString("local filesystem replace blocks wrk --done:\n")
			fmt.Fprintf(&b, "%s\n", replace.FormatIssueLine(top, issue))
			b.WriteString("resolve replace directives manually before running wrk --done")
			return errors.New(b.String())
		}

		// Only intra-repo offenders, default lenient mode: warn and proceed.
		fmt.Fprintln(os.Stderr, replace.FormatIssueLine(top, issue))
		fmt.Fprintln(os.Stderr, "local filesystem replace (intra-repo) - tolerated, remove before pushing:")
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
	// Branch includes the dep basename so that multiple distinct deps on the
	// same source branch (e.g. wrk --all-deps) do not collide on the branch
	// name and get spurious path suffixes. The printed path/name is unchanged.
	branch = basename + "-" + branchBase + "-" + date
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

func runCreate(workDir string, origWd string, targetDir string, taskDesc string, noCd bool) error {
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

	// Derive task slug if --task was set.
	// CLI rejects empty/whitespace task text when the flag is present with empty value.
	var slug string
	if taskDesc != "" {
		if strings.TrimSpace(taskDesc) == "" {
			return fmt.Errorf("wrk: task description must not be empty")
		}
		slug = slugify(taskDesc)
		if slug == "" {
			return fmt.Errorf("wrk: task description %q produces an empty slug", taskDesc)
		}
	}

	if targetDir != "" {
		return runCreateTargetDir(origWd, targetDir, checkoutRoot, mainRepo, basename, branchBase, pathToken, date, slug, noCd)
	}

	wrkHome, err := resolveWrkHome()
	if err != nil {
		return err
	}

	result, err := CreateDefaultWorktree(cwd, wrkHome, slug)
	if err != nil {
		return err
	}
	fmt.Println(result.Path)
	// Home-gate on shell process cwd (origWd), not source workDir.
	return writeFollowupCDIfCwdIsHome(noCd, origWd, result.Path)
}

// runCreateTargetDir handles wrk <dir> <target-dir>. A relative <target-dir> is
// resolved against origWd (the process/shell cwd), NOT the repo dir that Run
// chdir'd into.
func runCreateTargetDir(origWd, targetDir, checkoutRoot, mainRepo, basename, branchBase, pathToken, date, slug string, noCd bool) error {
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
			wtPath, branch := candidateNames(absTarget, basename, branchBase, pathToken, date, slug, suffix)
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
			return writeFollowupCDIfCwdIsHome(noCd, origWd, absPath)
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
	if slug != "" {
		branch = branch + "-" + slug
	}
	if err := createWorktree(checkoutRoot, wtPath, branch, branchExists(mainRepo, branch)); err != nil {
		return err
	}
	absPath, err := filepath.Abs(wtPath)
	if err != nil {
		return fmt.Errorf("resolve worktree path: %w", err)
	}
	fmt.Println(absPath)
	return writeFollowupCDIfCwdIsHome(noCd, origWd, absPath)
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
	cmd := gitCommand("-C", repo, "rev-parse", "--short=7", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --short=7 HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func sanitizeBranchToken(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func candidateNames(worktreesDir, basename, branchBase, pathToken, date, slug string, suffix int) (path, branch string) {
	name := fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
	if slug != "" {
		name = fmt.Sprintf("%s-%s", name, slug)
	}
	branch = branchBase + "-" + date
	if slug != "" {
		branch = branch + "-" + slug
	}
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
	cmd := gitCommand("-C", repo, "rev-parse", "--verify", "refs/heads/"+branch)
	return cmd.Run() == nil
}

func createWorktree(sourceDir, wtPath, branch string, branchPreExists bool) error {
	if !branchPreExists {
		cmd := gitCommand("-C", sourceDir, "worktree", "add", "-b", branch, wtPath)
		return runGitWorktreeAdd(cmd)
	}

	cmd := gitCommand("-C", sourceDir, "worktree", "add", "--no-checkout", wtPath)
	if err := runGitWorktreeAdd(cmd); err != nil {
		return err
	}

	checkout := gitCommand("-C", wtPath, "checkout", "--ignore-other-worktrees", branch)
	if out, err := checkout.CombinedOutput(); err != nil {
		_ = gitCommand("-C", sourceDir, "worktree", "remove", "--force", wtPath).Run()
		return fmt.Errorf("git checkout: %w\n%s", err, out)
	}
	return nil
}


// hasArg returns true if args contains the given flag.
func hasArg(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// slugify converts a task description into a path-safe slug.
// Rules: lowercase, non-letter-non-digit -> "-", collapse runs of "-",
// trim leading/trailing "-", truncate to 64 runes.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	s = b.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	runes := []rune(s)
	if len(runes) > 64 {
		s = string(runes[:64])
	}
	s = strings.Trim(s, "-")
	return s
}

// datePattern matches "-YYYY-MM-DD" in branch names for parsing wrk naming conventions.
var datePattern = regexp.MustCompile(`-(\d{4}-\d{2}-\d{2})`)

// parseBranchNaming extracts branchBase, date, slug, and suffix from a wrk-style
// branch name like "master-2026-07-01-fix-login-1". Returns an error if the
// branch doesn't contain a recognizable date pattern.
func parseBranchNaming(branch string) (branchBase, date, slug string, suffix int, err error) {
	loc := datePattern.FindStringSubmatchIndex(branch)
	if loc == nil {
		return "", "", "", 0, fmt.Errorf("no date pattern in branch name %q", branch)
	}
	branchBase = branch[:loc[0]]
	date = branch[loc[2]:loc[3]]
	tail := branch[loc[1]:] // includes leading "-"
	if tail == "" {
		return branchBase, date, "", 0, nil
	}
	tail = tail[1:] // strip leading "-"
	parts := strings.Split(tail, "-")
	last := parts[len(parts)-1]

	if n, convErr := strconv.Atoi(last); convErr == nil && n >= 0 && n < 100 {
		if len(parts) > 1 {
			slug = strings.Join(parts[:len(parts)-1], "-")
			suffix = n
		} else {
			suffix = n
		}
	} else {
		slug = tail
	}
	return branchBase, date, slug, suffix, nil
}

// runSetTask renames a linked worktree via git worktree move to include a new
// task slug in the directory and branch names. Requires TTY confirmation (or
// WRK_SET_TASK_CONFIRM=1 env var) before executing the move.
func runSetTask(workDir string, taskDesc string, assumeYes, noCd bool) error {
	if strings.TrimSpace(taskDesc) == "" {
		return fmt.Errorf("wrk: task description must not be empty")
	}
	newSlug := slugify(taskDesc)
	if newSlug == "" {
		return fmt.Errorf("wrk: task description %q produces an empty slug", taskDesc)
	}

	// Shell process cwd (inherited from interactive shell), not merely workDir.
	// Used after move to decide whether auto-cd is needed.
	shellCwd, _ := os.Getwd()

	cwd, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolve cwd: %w", err)
	}

	if !worktree.IsLinked(cwd) {
		return fmt.Errorf("wrk: --set-task must be run from inside a linked worktree")
	}

	branch, err := worktree.ReadBranch(cwd)
	if err != nil {
		return fmt.Errorf("read branch: %w", err)
	}

	branchBase, date, _, _, err := parseBranchNaming(branch)
	if err != nil {
		return fmt.Errorf("wrk: cannot parse branch name %q — is this a wrk worktree? (%w)", branch, err)
	}

	mainRepo, err := worktree.ResolveMainRepo(cwd)
	if err != nil {
		return fmt.Errorf("resolve main repo: %w", err)
	}

	basename := filepath.Base(mainRepo)
	pathToken := sanitizeBranchToken(branchBase)

	// Compute new names. We don't know the old suffix from the dir name alone,
	// so we derive it from the current dir basename. Find the wrk-style naming
	// by looking for the date pattern in the dir basename.
	curBase := filepath.Base(cwd)
	curLoc := datePattern.FindStringSubmatchIndex(curBase)
	if curLoc == nil {
		return fmt.Errorf("wrk: cannot parse directory name %q — is this a wrk worktree?", curBase)
	}
	curDate := curBase[curLoc[2]:curLoc[3]]
	curTail := curBase[curLoc[1]:]
	curSuffix := 0
	if curTail != "" {
		curTail = curTail[1:] // strip leading "-"
		// Remove the old slug (if any) from the tail to extract the suffix.
		// After date: [-slug][-N]. The suffix is at the very end if numeric.
		parts := strings.Split(curTail, "-")
		last := parts[len(parts)-1]
		if n, convErr := strconv.Atoi(last); convErr == nil && n >= 0 && n < 100 {
			curSuffix = n
		}
	}

	if curDate != date {
		return fmt.Errorf("wrk: date mismatch between branch (%s) and directory (%s)", date, curDate)
	}

	newDirName := fmt.Sprintf("%s-%s-%s", basename, pathToken, date)
	if newSlug != "" {
		newDirName = fmt.Sprintf("%s-%s", newDirName, newSlug)
	}
	if curSuffix > 0 {
		newDirName = fmt.Sprintf("%s-%d", newDirName, curSuffix)
	}

	newBranch := branchBase + "-" + date
	if newSlug != "" {
		newBranch = newBranch + "-" + newSlug
	}
	if curSuffix > 0 {
		newBranch = fmt.Sprintf("%s-%d", newBranch, curSuffix)
	}

	// If nothing changed, just report.
	if newDirName == curBase && newBranch == branch {
		fmt.Println("task unchanged")
		return nil
	}

	parentDir := filepath.Dir(cwd)
	newPath := filepath.Join(parentDir, newDirName)

	// Check if new path already exists
	if _, err := os.Stat(newPath); err == nil {
		return fmt.Errorf("wrk: target path %s already exists", newPath)
	}

	// Before renaming: discover nested linked worktrees under cwd so we can
	// update their gitdir metadata after the move.
	type nestedWT struct {
		oldPath string
		relPath string // relative to cwd
	}
	var nested []nestedWT
	repos, err := discoverStatusRepos(context.Background(), cwd)
	if err != nil {
		return fmt.Errorf("discover nested worktrees: %w", err)
	}
	cleanCwd := filepath.Clean(cwd)
	for _, repo := range repos {
		if repo.RepoType != scan_repo.RepoTypeWorktree {
			continue
		}
		if !worktree.IsLinked(repo.Path) {
			continue
		}
		if filepath.Clean(repo.Path) == cleanCwd {
			continue
		}
		rel, err := filepath.Rel(cwd, repo.Path)
		if err != nil {
			continue
		}
		nested = append(nested, nestedWT{oldPath: repo.Path, relPath: rel})
	}

	// TTY check (escape hatch for testing via WRK_SET_TASK_CONFIRM=1; -y bypasses)
	if !assumeYes && os.Getenv("WRK_SET_TASK_CONFIRM") != "1" {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			return fmt.Errorf("wrk: --set-task requires a terminal (tty)")
		}
		fmt.Printf("Rename worktree:\n  %s → %s\n  branch %s → %s\n", cwd, newPath, branch, newBranch)
		fmt.Print("Proceed? [Y/n] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer != "" && answer != "y" && answer != "Y" {
			return fmt.Errorf("wrk: --set-task aborted")
		}
	}

	// Execute git worktree move
	cmd := gitCommand("-C", mainRepo, "worktree", "move", cwd, newPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree move: %w\n%s", err, out)
	}

	// Also rename the branch
	branchCmd := gitCommand("-C", mainRepo, "branch", "-m", branch, newBranch)
	out, err = branchCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch rename: %w\n%s", err, out)
	}

	// Update gitdir metadata for nested worktrees that moved with the parent.
	// Each nested worktree's .git file says "gitdir: <mainRepo>/.git/worktrees/<name>",
	// and <mainRepo>/.git/worktrees/<name>/gitdir contains the old absolute path
	// back to the worktree. We rewrite it to the new path.
	for _, nw := range nested {
		newWtPath := filepath.Join(newPath, nw.relPath)
		gitFile := filepath.Join(newWtPath, ".git")
		data, err := os.ReadFile(gitFile)
		if err != nil {
			continue
		}
		s := strings.TrimSpace(string(data))
		const gitdirPrefix = "gitdir: "
		if !strings.HasPrefix(s, gitdirPrefix) {
			continue
		}
		gitdirBase := strings.TrimSpace(s[len(gitdirPrefix):])
		gitdirFile := filepath.Join(gitdirBase, "gitdir")
		newGitdirContent := filepath.Join(newWtPath, ".git") + "\n"
		_ = os.WriteFile(gitdirFile, []byte(newGitdirContent), 0644)
	}

	if err := rewriteConsumerReplacePaths(cwd, newPath); err != nil {
		return fmt.Errorf("rewrite go.mod replace paths: %w", err)
	}

	fmt.Println(newPath)
	// Write follow-up cd only if shell cwd is gone (user was inside the moved
	// worktree). Surviving sibling/main stays put.
	return writeFollowupCDIfCwdMissing(noCd, shellCwd, newPath)
}
