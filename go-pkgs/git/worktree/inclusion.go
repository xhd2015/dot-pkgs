package worktree

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/gitops/git"
)

// InclusionResult describes whether a worktree HEAD is contained in main HEAD.
type InclusionResult struct {
	Included bool
	Relation string // "same", "ancestor", "ahead", "diverged"
}

// HeadIncludedInMain checks whether the worktree HEAD is an ancestor of (or equal
// to) the main repository HEAD.
func HeadIncludedInMain(mainRepo, worktreePath string) (*InclusionResult, error) {
	ref, err := worktreeRef(worktreePath)
	if err != nil {
		return nil, err
	}

	result, err := git.CompareBranches(mainRepo, ref, "HEAD")
	if err != nil {
		return nil, err
	}

	switch result.Relation {
	case git.BranchRelationSame:
		return &InclusionResult{Included: true, Relation: "same"}, nil
	case git.BranchRelationAIsAncestorOfB:
		return &InclusionResult{Included: true, Relation: "ancestor"}, nil
	case git.BranchRelationBIsAncestorOfA:
		return &InclusionResult{Included: false, Relation: "ahead"}, nil
	case git.BranchRelationDiverged:
		return &InclusionResult{Included: false, Relation: "diverged"}, nil
	default:
		return nil, fmt.Errorf("unknown branch relation")
	}
}

func worktreeRef(worktreePath string) (string, error) {
	branch, err := ReadBranch(worktreePath)
	if err != nil {
		return "", err
	}
	if branch != "HEAD" {
		return branch, nil
	}
	cmd := exec.Command("git", "-C", worktreePath, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}