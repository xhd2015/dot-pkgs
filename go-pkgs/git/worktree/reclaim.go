package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ActionReclaimed = "reclaimed"
	ActionDryRun    = "dry-run"
	ActionSkipped   = "skipped"
	ActionError     = "error"
)

// ReclaimOptions configures a reclaim operation.
type ReclaimOptions struct {
	Cwd    string // working directory for resolving main repo when All is true
	Path   string // single linked worktree path
	All    bool
	DryRun bool
	// OnOutcome is called immediately after each worktree is processed.
	// Used by CLI callers to stream progress to stdout.
	OnOutcome func(ReclaimOutcome)
}

// ReclaimOutcome describes the result for one worktree.
type ReclaimOutcome struct {
	Path   string
	Action string // reclaimed, dry-run, skipped, error
	Reason string
}

// Reclaim removes reclaimable linked worktrees.
//
// For single-path mode, returns an error when the target is not reclaimable or
// removal fails. For --all mode, non-reclaimable worktrees are skipped and only
// removal failures produce an error.
func Reclaim(opts ReclaimOptions) ([]ReclaimOutcome, error) {
	if opts.All && opts.Path != "" {
		return nil, fmt.Errorf("cannot use --all with a worktree path")
	}
	if !opts.All && opts.Path == "" {
		return nil, fmt.Errorf("requires <worktree-dir> or --all")
	}

	if opts.All {
		return reclaimAll(opts)
	}
	return reclaimSingle(opts)
}

func reclaimAll(opts ReclaimOptions) ([]ReclaimOutcome, error) {
	cwd := opts.Cwd
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
	}

	mainRepo, err := ResolveMainRepo(cwd)
	if err != nil {
		return nil, err
	}

	linked, err := ListLinked(mainRepo)
	if err != nil {
		return nil, err
	}

	var outcomes []ReclaimOutcome
	var removalErr error
	for _, entry := range linked {
		outcome, err := reclaimOne(mainRepo, entry.Path, entry.Branch, opts.DryRun, true)
		emitOutcome(opts, outcome)
		outcomes = append(outcomes, outcome)
		if err != nil {
			removalErr = err
		}
	}
	return outcomes, removalErr
}

func reclaimSingle(opts ReclaimOptions) ([]ReclaimOutcome, error) {
	path, err := filepath.Abs(opts.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return reclaimSingleDead(opts, path)
		}
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	if !IsLinked(path) {
		return nil, fmt.Errorf("%s is not a linked worktree", path)
	}

	mainRepo, err := ReadMainRepo(path)
	if err != nil {
		return nil, err
	}

	branch, err := ReadBranch(path)
	if err != nil {
		return nil, err
	}

	outcome, err := reclaimOne(mainRepo, path, branch, opts.DryRun, false)
	emitOutcome(opts, outcome)
	return []ReclaimOutcome{outcome}, err
}

func reclaimSingleDead(opts ReclaimOptions, path string) ([]ReclaimOutcome, error) {
	cwd := opts.Cwd
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
	}

	mainRepo, err := ResolveMainRepo(cwd)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", path)
	}

	linked, err := ListLinked(mainRepo)
	if err != nil {
		return nil, err
	}

	for _, entry := range linked {
		entryPath, err := filepath.Abs(entry.Path)
		if err != nil {
			continue
		}
		if samePath(entryPath, path) {
			outcome, err := reclaimOne(mainRepo, entryPath, entry.Branch, opts.DryRun, false)
			emitOutcome(opts, outcome)
			return []ReclaimOutcome{outcome}, err
		}
	}

	return nil, fmt.Errorf("path does not exist: %s", path)
}

func emitOutcome(opts ReclaimOptions, outcome ReclaimOutcome) {
	if opts.OnOutcome != nil {
		opts.OnOutcome(outcome)
	}
}

func reclaimOne(mainRepo, path, branch string, dryRun, skipNonReclaimable bool) (ReclaimOutcome, error) {
	if IsDead(path) {
		if dryRun {
			return ReclaimOutcome{Path: path, Action: ActionDryRun, Reason: "dead"}, nil
		}
		if err := removeWorktree(mainRepo, path, true); err != nil {
			return ReclaimOutcome{Path: path, Action: ActionError, Reason: err.Error()}, err
		}
		if branch != "" && branch != "HEAD" {
			if err := deleteBranch(mainRepo, branch); err != nil {
				return ReclaimOutcome{Path: path, Action: ActionError, Reason: err.Error()}, err
			}
		}
		return ReclaimOutcome{Path: path, Action: ActionReclaimed, Reason: "dead"}, nil
	}

	if err := IsClean(path); err != nil {
		if strings.Contains(err.Error(), "uncommitted changes") {
			reason := "uncommitted changes"
			if skipNonReclaimable {
				return ReclaimOutcome{Path: path, Action: ActionSkipped, Reason: reason}, nil
			}
			return ReclaimOutcome{Path: path, Action: ActionError, Reason: reason}, fmt.Errorf("worktree %s is not clean: %s", path, reason)
		}
		reason := err.Error()
		if skipNonReclaimable {
			return ReclaimOutcome{Path: path, Action: ActionSkipped, Reason: reason}, nil
		}
		return ReclaimOutcome{Path: path, Action: ActionError, Reason: reason}, err
	}

	inclusion, err := HeadIncludedInMain(mainRepo, path)
	if err != nil {
		return ReclaimOutcome{Path: path, Action: ActionError, Reason: err.Error()}, err
	}
	if !inclusion.Included {
		reason := inclusionReason(inclusion.Relation)
		if skipNonReclaimable {
			return ReclaimOutcome{Path: path, Action: ActionSkipped, Reason: reason}, nil
		}
		return ReclaimOutcome{Path: path, Action: ActionError, Reason: reason}, fmt.Errorf("worktree HEAD is not included in main HEAD: %s", reason)
	}

	if dryRun {
		return ReclaimOutcome{Path: path, Action: ActionDryRun}, nil
	}

	if err := removeWorktree(mainRepo, path, false); err != nil {
		return ReclaimOutcome{Path: path, Action: ActionError, Reason: err.Error()}, err
	}

	if branch != "" && branch != "HEAD" {
		if err := deleteBranch(mainRepo, branch); err != nil {
			return ReclaimOutcome{Path: path, Action: ActionError, Reason: err.Error()}, err
		}
	}

	return ReclaimOutcome{Path: path, Action: ActionReclaimed}, nil
}

func inclusionReason(relation string) string {
	switch relation {
	case "ahead":
		return "HEAD is ahead of main HEAD"
	case "diverged":
		return "diverged from main HEAD"
	default:
		return "HEAD not included in main HEAD"
	}
}

func removeWorktree(mainRepo, path string, force bool) error {
	args := []string{"-C", mainRepo, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	cmd := exec.Command("git", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git worktree remove: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func deleteBranch(mainRepo, branch string) error {
	cmd := exec.Command("git", "-C", mainRepo, "branch", "-D", branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git branch -D %s: %w\n%s", branch, err, strings.TrimSpace(string(out)))
	}
	return nil
}