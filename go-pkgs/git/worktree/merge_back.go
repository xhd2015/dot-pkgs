package worktree

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
	"github.com/xhd2015/gitops/git"
)

// ErrConfirmationRequired is returned when merge/rebase mutations need user
// confirmation but stdin is not interactive.
var ErrConfirmationRequired = errors.New("stdin is not a terminal; cannot prompt for confirmation")

// PlannedCommand is one git command MergeBack would execute.
type PlannedCommand struct {
	Dir  string   // -C argument
	Args []string // git subcommand args (without the "git" prefix)
}

func (c PlannedCommand) String() string {
	return fmt.Sprintf("git -C %s %s", c.Dir, strings.Join(c.Args, " "))
}

// MergeBackPlan describes the git operations MergeBack would perform.
type MergeBackPlan struct {
	Relation     string // "same", "ancestor", "ahead", "diverged"
	Branch       string
	Commands     []PlannedCommand
	NeedsConfirm bool
	TargetLabel  string
}

// MergeBackOptions configures a merge-back operation.
type MergeBackOptions struct {
	SourcePath string
	TargetPath string
	DryRun     bool
	Remove     bool
	// Confirm is called before executing when plan.NeedsConfirm.
	// Must return (true, nil) to proceed. (false, nil) aborts cleanly.
	// nil + NeedsConfirm → ErrConfirmationRequired (non-interactive).
	Confirm func(plan MergeBackPlan) (bool, error)
}

// MergeBackResult describes the outcome of a merge-back operation.
type MergeBackResult struct {
	SourcePath string
	TargetPath string
	Branch     string
	Relation   string
	Action     string // "noop", "merged", "rebased-and-merged", "removed", "dry-run", "aborted"
	Message    string
}

// MergeBack merges a linked worktree branch into a target checkout.
func MergeBack(opts MergeBackOptions) (*MergeBackResult, error) {
	sourceAbs, err := filepath.Abs(opts.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("resolve source path: %w", err)
	}

	if !IsLinked(sourceAbs) {
		return nil, fmt.Errorf("%s is not a linked worktree", sourceAbs)
	}

	mainRepo, err := ReadMainRepo(sourceAbs)
	if err != nil {
		return nil, err
	}

	targetAbs := mainRepo
	if opts.TargetPath != "" {
		targetAbs, err = filepath.Abs(opts.TargetPath)
		if err != nil {
			return nil, fmt.Errorf("resolve target path: %w", err)
		}
	}

	if samePath(sourceAbs, targetAbs) {
		return nil, fmt.Errorf("source and target are the same worktree")
	}

	targetMain, err := ResolveMainRepo(targetAbs)
	if err != nil {
		return nil, err
	}
	if !samePath(mainRepo, targetMain) {
		return nil, fmt.Errorf("target does not share the same main repository")
	}

	branch, err := ReadBranch(sourceAbs)
	if err != nil {
		return nil, err
	}

	compareRef, err := worktreeRef(sourceAbs)
	if err != nil {
		return nil, err
	}
	mergeRef := compareRef
	if branch != "HEAD" {
		mergeRef = branch
	}

	compare, err := git.CompareBranches(targetAbs, compareRef, "HEAD")
	if err != nil {
		return nil, err
	}

	relation, included := relationFromCompare(compare.Relation)

	result := &MergeBackResult{
		SourcePath: sourceAbs,
		TargetPath: targetAbs,
		Branch:     branch,
		Relation:   relation,
	}

	// Dirty check: only required when --rm (Remove: true) because the worktree
	// will be deleted. When Remove is false, the worktree stays — dirtiness is
	// irrelevant. Diverged + dirty + !Remove uses tmp worktree for rebase.
	dirtyErr := IsClean(sourceAbs)
	dirty := dirtyErr != nil
	if dirty && opts.Remove {
		return nil, dirtyErr
	}

	plan, err := buildMergeBackPlan(mergeBackPlanInput{
		relation:    relation,
		branch:      mergeRef,
		sourcePath:  sourceAbs,
		targetPath:  targetAbs,
		mainRepo:    mainRepo,
		remove:      opts.Remove,
		targetLabel: targetLabel(targetAbs),
	})
	if err != nil {
		return nil, err
	}

	if included {
		if !opts.Remove {
			result.Action = "noop"
			result.Message = fmt.Sprintf("branch %s is already included in %s", branch, plan.TargetLabel)
			return result, nil
		}
		if opts.DryRun {
			return printDryRun(result, plan)
		}
		if err := executeRemove(plan, sourceAbs, mainRepo, branch); err != nil {
			return nil, err
		}
		result.Action = "removed"
		result.Message = fmt.Sprintf("worktree removed: %s [branch: %s deleted]", sourceAbs, branch)
		return result, nil
	}

	// For diverged + dirty + !remove: use tmp worktree for rebase
	if relation == "diverged" && dirty && !opts.Remove {
		return mergeBackViaTmpWorktree(opts, result, sourceAbs, mainRepo, targetAbs, branch, mergeRef)
	}

	if opts.DryRun {
		return printDryRun(result, plan)
	}

	if plan.NeedsConfirm {
		if opts.Confirm == nil {
			return nil, ErrConfirmationRequired
		}
		confirmed, err := opts.Confirm(plan)
		if err != nil {
			return nil, err
		}
		if !confirmed {
			result.Action = "aborted"
			result.Message = "merge-back aborted"
			return result, nil
		}
	}

	if err := executePlan(plan); err != nil {
		return nil, err
	}

	switch relation {
	case "ahead":
		result.Action = "merged"
		result.Message = fmt.Sprintf("merged branch %s into %s", branch, plan.TargetLabel)
	case "diverged":
		result.Action = "rebased-and-merged"
		result.Message = fmt.Sprintf("rebased and merged branch %s into %s", branch, plan.TargetLabel)
	}
	return result, nil
}

// mergeBackViaTmpWorktree handles the diverged+dirty+!Remove case by creating
// a temporary worktree, rebasing there, merging, and cleaning up.
func mergeBackViaTmpWorktree(
	opts MergeBackOptions,
	result *MergeBackResult,
	sourceAbs, mainRepo, targetAbs, branch, mergeRef string,
) (*MergeBackResult, error) {
	targetHEAD, err := revParseCommit(targetAbs, "HEAD")
	if err != nil {
		return nil, err
	}

	targetLabelStr := targetLabel(targetAbs)

	tmpPath, tmpBranch, err := createTmpWorktree(mainRepo, branch, mergeRef)
	if err != nil {
		return nil, fmt.Errorf("create tmp worktree: %w", err)
	}
	defer cleanupTmpWorktree(mainRepo, tmpPath, tmpBranch)

	// Build tmp plan: rebase in tmp worktree, merge tmp branch
	tmpPlan := MergeBackPlan{
		Relation:     "diverged",
		Branch:       branch,
		TargetLabel:  targetLabelStr,
		NeedsConfirm: true,
		Commands: []PlannedCommand{
			{Dir: tmpPath, Args: []string{"rebase", targetHEAD}},
			{Dir: targetAbs, Args: []string{"merge", "--ff-only", tmpBranch}},
		},
	}

	if opts.DryRun {
		return printDryRun(result, tmpPlan)
	}

	if tmpPlan.NeedsConfirm {
		if opts.Confirm == nil {
			return nil, ErrConfirmationRequired
		}
		confirmed, err := opts.Confirm(tmpPlan)
		if err != nil {
			return nil, err
		}
		if !confirmed {
			result.Action = "aborted"
			result.Message = "merge-back aborted"
			return result, nil
		}
	}

	if err := executePlan(tmpPlan); err != nil {
		return nil, err
	}

	// Stash user's dirty changes from source, test them against the
	// rebased HEAD in tmp, then migrate them back to source after
	// the branch ref is updated.
	if err := migrateDirtyChanges(sourceAbs, mainRepo, tmpPath, branch); err != nil {
		return nil, err
	}

	// Force-update the source branch to the rebased position.
	tmpCommit, err := revParseCommit(mainRepo, tmpBranch)
	if err != nil {
		return nil, fmt.Errorf("resolve tmp branch commit: %w", err)
	}
	if err := runGit(mainRepo, "update-ref", "refs/heads/"+branch, tmpCommit); err != nil {
		return nil, fmt.Errorf("update source branch ref: %w", err)
	}

	// Reset source worktree to clean rebased HEAD.
	if err := runGit(sourceAbs, "reset", "--hard", "HEAD"); err != nil {
		return nil, fmt.Errorf("sync source worktree: %w", err)
	}
	if err := runGit(sourceAbs, "clean", "-fd"); err != nil {
		return nil, fmt.Errorf("clean source worktree: %w", err)
	}

	// Pop the migrated stash back onto source.
	if err := runGit(sourceAbs, "stash", "pop"); err != nil {
		return nil, fmt.Errorf("restore dirty changes to source: %w", err)
	}

	result.Action = "rebased-and-merged"
	result.Message = fmt.Sprintf("rebased and merged branch %s into %s", branch, tmpPlan.TargetLabel)
	return result, nil
}

// migrateDirtyChanges stashes dirty tracked+untracked changes from source,
// applies them to tmp (the rebased worktree) to check for conflicts, then
// re-stashes them on tmp. If apply-on-tmp conflicts, the source stash is
// preserved so the user can manually resolve on source.
func migrateDirtyChanges(sourceAbs, mainRepo, tmpPath, branch string) error {
	// git stash push includes untracked files (-u) but not ignored.
	// On clean worktrees this is a no-op (empty stash).
	// We push first, then check if anything was stashed.
	if err := runGit(sourceAbs, "stash", "push", "-u", "-m", "wrk-merge-back"); err != nil {
		return fmt.Errorf("stash dirty changes: %w", err)
	}

	// Check if stash was actually created (dirty worktree may have been clean).
	stashList, err := gitOutput(sourceAbs, "stash", "list", "-1")
	if err != nil {
		return err
	}
	if !strings.Contains(stashList, "wrk-merge-back") {
		// Source was clean — nothing to migrate.
		return nil
	}

	// Apply the stash to the tmp worktree (rebased HEAD) to detect conflicts.
	if err := runGit(tmpPath, "stash", "apply"); err != nil {
		// Conflict: working tree has conflict markers, stash survives.
		// Clean up tmp, pop stash back on source, report error.
		runGit(tmpPath, "reset", "--hard", "HEAD")
		runGit(tmpPath, "clean", "-fd")
		runGit(sourceAbs, "stash", "pop")
		return fmt.Errorf("dirty changes conflict with rebase — resolve manually, then retry: %w", err)
	}

	// Applied cleanly on tmp. Drop the original stash (we'll re-stash from tmp).
	if err := runGit(sourceAbs, "stash", "drop"); err != nil {
		return fmt.Errorf("drop original stash: %w", err)
	}

	// Re-stash from tmp — this captures the user's changes relative to
	// the rebased HEAD. After source branch is update-ref'd and reset,
	// we'll pop this stash on source.
	if err := runGit(tmpPath, "stash", "push", "-u", "-m", "wrk-merge-back"); err != nil {
		return fmt.Errorf("re-stash from tmp: %w", err)
	}

	return nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return string(out), fmt.Errorf("git %v: %w\n%s", args, err, exitErr.Stderr)
		}
		return string(out), err
	}
	return strings.TrimSpace(string(out)), nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v: %w\n%s", args, err, combinedOutput(out))
	}
	return nil
}

func combinedOutput(out []byte) string { return strings.TrimSpace(string(out)) }

func createTmpWorktree(mainRepo, sourceBranch, mergeRef string) (tmpPath, tmpBranch string, err error) {
	worktreesDir, err := resolveWrkWorktreesDir()
	if err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(worktreesDir, 0755); err != nil {
		return "", "", err
	}

	basename := filepath.Base(mainRepo)
	pathToken := sanitizeBranchToken(sourceBranch)
	date := resolveDate()

	for suffix := 0; suffix < 100; suffix++ {
		// tmp branch: source-branch-tmp-rebase-<random8>
		candidateBranch := sourceBranch + "-tmp-rebase-" + random8()

		// tmp path with optional suffix for directory collision
		name := fmt.Sprintf("%s-%s-%s-tmp-rebase", basename, pathToken, date)
		if suffix > 0 {
			name = fmt.Sprintf("%s-%d", name, suffix)
		}
		candidatePath := filepath.Join(worktreesDir, name)

		if _, statErr := os.Stat(candidatePath); statErr == nil {
			continue // directory already exists, try next suffix
		}

		// Create tmp worktree with the branch
		if err := runGit(mainRepo, "worktree", "add", "-b", candidateBranch, candidatePath, mergeRef); err != nil {
			// Branch collision? Try again with a new random branch name.
			continue
		}

		return candidatePath, candidateBranch, nil
	}

	return "", "", fmt.Errorf("could not create tmp worktree after 100 attempts")
}

func cleanupTmpWorktree(mainRepo, tmpPath, tmpBranch string) {
	exec.Command("git", "-C", mainRepo, "worktree", "remove", "--force", tmpPath).Run()
	exec.Command("git", "-C", mainRepo, "branch", "-D", tmpBranch).Run()
}

func resolveWrkWorktreesDir() (string, error) {
	if v := os.Getenv("WRK_HOME"); v != "" {
		return filepath.Join(pathfmt.Expand(v), "worktrees"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".wrk", "worktrees"), nil
}

func resolveDate() string {
	if v := os.Getenv("WRK_DATE"); v != "" {
		return v
	}
	return time.Now().Format("2006-01-02")
}

func sanitizeBranchToken(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func random8() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type mergeBackPlanInput struct {
	relation    string
	branch      string
	sourcePath  string
	targetPath  string
	mainRepo    string
	remove      bool
	targetLabel string
}

func relationFromCompare(relation git.BranchRelation) (string, bool) {
	switch relation {
	case git.BranchRelationSame:
		return "same", true
	case git.BranchRelationAIsAncestorOfB:
		return "ancestor", true
	case git.BranchRelationBIsAncestorOfA:
		return "ahead", false
	case git.BranchRelationDiverged:
		return "diverged", false
	default:
		return "unknown", false
	}
}

func buildMergeBackPlan(in mergeBackPlanInput) (MergeBackPlan, error) {
	plan := MergeBackPlan{
		Relation:     in.relation,
		Branch:       in.branch,
		TargetLabel:  in.targetLabel,
		NeedsConfirm: in.relation == "ahead" || in.relation == "diverged",
	}

	switch in.relation {
	case "ahead":
		plan.Commands = append(plan.Commands, PlannedCommand{
			Dir:  in.targetPath,
			Args: []string{"merge", "--ff-only", in.branch},
		})
	case "diverged":
		targetHEAD, err := revParseCommit(in.targetPath, "HEAD")
		if err != nil {
			return plan, err
		}
		plan.Commands = append(plan.Commands, PlannedCommand{
			Dir:  in.sourcePath,
			Args: []string{"rebase", targetHEAD},
		})
		plan.Commands = append(plan.Commands, PlannedCommand{
			Dir:  in.targetPath,
			Args: []string{"merge", "--ff-only", in.branch},
		})
	}

	if in.remove {
		plan.Commands = append(plan.Commands,
			PlannedCommand{Dir: in.mainRepo, Args: []string{"worktree", "remove", in.sourcePath}},
			PlannedCommand{Dir: in.mainRepo, Args: []string{"branch", "-D", in.branch}},
		)
	}

	return plan, nil
}

// displayPath shortens paths for CLI output consistently with doctest fixtures
// that build paths via filepath.Join (macOS /var vs /private/var).
func displayPath(p string) string {
	p = filepath.Clean(p)
	if strings.HasPrefix(p, "/private/var/") {
		p = "/var/" + strings.TrimPrefix(p, "/private/var/")
	}
	return pathfmt.Short(p)
}

func targetLabel(targetAbs string) string {
	branch, err := ReadBranch(targetAbs)
	if err != nil || branch == "" || branch == "HEAD" {
		return filepath.Base(targetAbs)
	}
	return branch
}

func plannedCommandComment(cmd PlannedCommand, plan MergeBackPlan) string {
	if len(cmd.Args) == 0 {
		return ""
	}
	switch cmd.Args[0] {
	case "merge":
		return fmt.Sprintf("# %s: fast forward", plan.TargetLabel)
	case "rebase":
		return fmt.Sprintf("# %s: rebase onto %s", plan.Branch, plan.TargetLabel)
	case "worktree":
		if len(cmd.Args) >= 2 && cmd.Args[1] == "remove" {
			return "# worktree: remove"
		}
	case "branch":
		if len(cmd.Args) >= 2 && cmd.Args[1] == "-D" {
			return "# worktree branch: drop"
		}
	}
	return ""
}

func formatPlannedCommandForDisplay(cmd PlannedCommand) string {
	dir := displayPath(cmd.Dir)
	args := append([]string(nil), cmd.Args...)
	if len(args) >= 3 && args[0] == "worktree" && args[1] == "remove" {
		args[2] = displayPath(args[2])
	}
	if len(args) >= 2 && args[0] == "rebase" {
		args = args[:1]
	}
	return fmt.Sprintf("git -C %s %s", dir, strings.Join(args, " "))
}

// WritePlannedCommandsDisplay writes indented comment + command lines for a plan.
func WritePlannedCommandsDisplay(b *strings.Builder, plan MergeBackPlan) {
	for _, cmd := range plan.Commands {
		if comment := plannedCommandComment(cmd, plan); comment != "" {
			fmt.Fprintf(b, "  %s\n", comment)
		}
		fmt.Fprintf(b, "  %s\n", formatPlannedCommandForDisplay(cmd))
	}
}

func printDryRun(result *MergeBackResult, plan MergeBackPlan) (*MergeBackResult, error) {
	var b strings.Builder
	b.WriteByte('\n')
	WritePlannedCommandsDisplay(&b, plan)
	fmt.Print(strings.TrimSuffix(b.String(), "\n"))
	result.Action = "dry-run"
	result.Message = "dry-run: planned commands listed"
	return result, nil
}

func executePlan(plan MergeBackPlan) error {
	for _, cmd := range plan.Commands {
		if err := runPlannedCommand(cmd, plan.Relation); err != nil {
			return err
		}
	}
	return nil
}

func runPlannedCommand(cmd PlannedCommand, relation string) error {
	gitCmd := exec.Command("git", append([]string{"-C", cmd.Dir}, cmd.Args...)...)
	out, err := gitCmd.CombinedOutput()
	if err != nil {
		if relation == "diverged" && len(cmd.Args) > 0 && cmd.Args[0] == "rebase" {
			abort := exec.Command("git", "-C", cmd.Dir, "rebase", "--abort")
			abort.Run()
			return fmt.Errorf("rebase conflict: %w\n%s", err, strings.TrimSpace(string(out)))
		}
		return fmt.Errorf("git %s: %w\n%s", strings.Join(cmd.Args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}

func executeRemove(plan MergeBackPlan, sourcePath, mainRepo, branch string) error {
	removeCmds := []PlannedCommand{
		{Dir: mainRepo, Args: []string{"worktree", "remove", sourcePath}},
	}
	if branch != "" && branch != "HEAD" {
		removeCmds = append(removeCmds, PlannedCommand{Dir: mainRepo, Args: []string{"branch", "-D", branch}})
	}
	for _, cmd := range removeCmds {
		if err := runPlannedCommand(cmd, plan.Relation); err != nil {
			return err
		}
	}
	return nil
}

func revParseCommit(dir, ref string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--verify", ref)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --verify %s: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}