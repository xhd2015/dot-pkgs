package worktree

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/cmd"
)

// errNoRemoteSync means the checkout has no configured upstream and no origin
// remote, so remote main-sync cannot run. Callers treat this as skip.
var errNoRemoteSync = errors.New("no upstream and no origin remote")

// ResolveBranchUpstream returns the remote and remote branch for repo's branch.
// Prefer branch.<name>.remote + branch.<name>.merge; else origin + branch when
// the origin remote exists.
func ResolveBranchUpstream(repo, branch string) (remote, remoteBranch string, err error) {
	branch = strings.TrimSpace(branch)
	if branch == "" || branch == "HEAD" {
		return "", "", fmt.Errorf("detached HEAD (not on a named branch)")
	}

	ctx := context.Background()
	remoteOut, ok, remErr := cmd.RunOptional(ctx, repo, "config", "--get", "branch."+branch+".remote")
	if remErr != nil {
		return "", "", remErr
	}
	if ok {
		remote = strings.TrimSpace(remoteOut)
	}
	if remote != "" {
		mergeOut, mergeOK, mergeErr := cmd.RunOptional(ctx, repo, "config", "--get", "branch."+branch+".merge")
		if mergeErr != nil {
			return "", "", mergeErr
		}
		if mergeOK {
			merge := strings.TrimSpace(mergeOut)
			remoteBranch = strings.TrimPrefix(merge, "refs/heads/")
		}
		if remoteBranch == "" {
			remoteBranch = branch
		}
		return remote, remoteBranch, nil
	}

	if _, originOK, originErr := cmd.RunOptional(ctx, repo, "remote", "get-url", "origin"); originErr != nil || !originOK {
		return "", "", errNoRemoteSync
	}
	return "origin", branch, nil
}

// planMainRemoteSync builds fetch + rebase commands to bring main onto its
// upstream tip. Returns errNoRemoteSync when no remote can be resolved.
func planMainRemoteSync(mainRepo string) ([]PlannedCommand, string, error) {
	branch, err := ReadBranch(mainRepo)
	if err != nil {
		return nil, "", err
	}
	remote, remoteBranch, err := ResolveBranchUpstream(mainRepo, branch)
	if err != nil {
		return nil, "", err
	}
	upstreamRef := remote + "/" + remoteBranch
	cmds := []PlannedCommand{
		{Dir: mainRepo, Args: []string{"fetch", remote, remoteBranch}},
		{Dir: mainRepo, Args: []string{"rebase", upstreamRef}},
	}
	return cmds, upstreamRef, nil
}
