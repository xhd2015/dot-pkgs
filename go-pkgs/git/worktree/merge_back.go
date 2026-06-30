package worktree

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

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

	if err := IsClean(sourceAbs); err != nil {
		return nil, err
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

	result := &MergeBackResult{
		SourcePath: sourceAbs,
		TargetPath: targetAbs,
		Branch:     branch,
		Relation:   relation,
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