package githook

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/xhd2015/gitops/gitwrite"
)

// UnstageFiles removes paths from the git index while leaving the working
// tree untouched, so hooks with --auto-unstage can unstage detected files
// without failing on repos that have no commits yet.
//
// With at least one commit it delegates to gitwrite.RestoreStaged: each
// index entry resets to HEAD, so modified tracked files stay tracked. On a
// fresh repo HEAD does not resolve and there is nothing to restore to; every
// staged path is new there, so removing it from the index with
// "git rm --cached" is exactly "unstage".
func UnstageFiles(paths ...string) error {
	if len(paths) == 0 {
		return nil
	}
	hasHead, err := repoHasHead()
	if err != nil {
		return err
	}
	if hasHead {
		return gitwrite.RestoreStaged(".", paths...)
	}
	args := append([]string{"rm", "--cached", "--quiet", "--"}, paths...)
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git rm --cached failed: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// repoHasHead reports whether HEAD resolves to a commit. An unborn branch
// (a repo with zero commits) reports false; rev-parse -q exits 1 without
// output in that case, which GitOptionalOutput reports as ok=false. Other
// git failures (not a repo, git missing) propagate as errors.
func repoHasHead() (bool, error) {
	_, ok, err := GitOptionalOutput("rev-parse", "--verify", "-q", "HEAD")
	if err != nil {
		return false, err
	}
	return ok, nil
}
