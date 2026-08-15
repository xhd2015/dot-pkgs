# Scenario

**Feature**: stash apply on tmp detects file-level modify/modify conflict

```
# Feature changed file-a.txt. Main changed file-b.txt. Rebase succeeds (different files).
# User dirtied file-b.txt with different content. stash apply on rebased HEAD conflicts.
dirty feat -> stash push -> rebase clean (different files) -> stash apply -> content conflict -> reject
```

## Steps

1. Feature already committed change to file-a.txt (from diverged/SETUP.md).
2. On main: modify file-b.txt (new file added to common ancestor isn't possible, use feature-only.txt — wait that's already the feature-only change).

Let me use README.md + another unrelated file.

Actually the diverged/SETUP already creates the base: feature committed `feature-only.txt`, main committed `main-only.txt`. Rebase will succeed (different files). Now user dirties one of the files the rebase result will have differently... but the rebase result has `main-only.txt` from main's commit AND `feature-only.txt` from feature's commit. Neither file existed in common ancestor. The rebase applies feature's commit on top of main, so both files exist in rebased HEAD.

If user dirties `main-only.txt` (which main added), the stash apply compares against old HEAD which has `main-only.txt` with OLD content? No — feature's old HEAD doesn't have `main-only.txt` at all. So stash would apply clean (file didn't exist before).

The right approach: use `README.md` — exists on common ancestor, feature committed a change, main committed a change → rebase would conflict on README.md. That's wrong.

OK, simplest approach for content conflict: feature changes file-a (from diverged/SETUP), main does NOT touch a shared file. User dirties `feature-only.txt` (which feature committed). After rebase, `feature-only.txt` has feature's committed content, user's working tree has different content. Stash apply → content conflict because the base (committed `feature-only.txt`) vs rebased (also `feature-only.txt`, same) actually NO conflict because content is same...

I'm overcomplicating this. Let me just use a scenario where:
- Common ancestor has `README.md` with content "init"
- Feature commits a change to `README.md` → "feature line"
- Main does NOT touch `README.md` → rebase succeeds (feature's change applies on top of main cleanly)
- User dirties `README.md` → "user dirty content"  
- After rebase, `README.md` = "feature line" (from feature's commit)
- Stash has `README.md` = "user dirty content" (vs old HEAD which was "feature line")
- Actually no conflict — same file, stash apply just overwrites. That's a clean apply!

The problem is I can't easily trigger a stash conflict without a rebase conflict when both branches modify the same file. Unless:

- Feature commits changes to `a.txt`
- Main commits changes to `b.txt`  
- Rebase succeeds (different files)
- User dirties `a.txt` BEFORE feature committed it (i.e., `a.txt` exists on common ancestor, feature committed a change, user has a different uncommitted change)
- After rebase: `a.txt` has feature's committed content. User's working tree had different content. Stash says "restore `a.txt` to user's content vs old HEAD which was feature's content". Actually — the old HEAD IS feature's commit. So base `a.txt` = feature's committed content. Rebased `a.txt` = feature's committed content. It's the SAME. No conflict!

This is fundamentally hard because the tmp worktree always rebases the branch commits, and the stash compares against the branch's OLD HEAD. If the feature's own commit didn't change a file, and the rebase didn't change it (main didn't touch it), then the file is unchanged after rebase → no stash conflict.

The ONLY way to get a stash conflict is if the rebase CHANGED a file that the user also changed. But if the rebase changed it, that means main touched it. And if main touched it AND user changed it, git stash apply sees: old base = feature HEAD content, new base = rebased (includes main's change), user's stash = user content. The 3-way merge: old vs new is main's change, old vs user is user's change, same file → CONFLICT.

For this to work without rebase conflict: feature branch must NOT have a commit touching that file. Only main does. And user dirties it. Then rebase succeeds (feature didn't touch it), stash apply conflicts.

So the setup: common ancestor has `file.txt`. Feature branch has NO commits touching `file.txt`. Main has a commit modifying `file.txt`. User dirties `file.txt`. Rebase: feature's commits (none touching `file.txt`) apply cleanly on main → `file.txt` is main's version. Stash: user's version on top of main's version → 3-way merge: base is common ancestor's `file.txt`, new is main's `file.txt`, user's is user's `file.txt` → content conflict (if changed same area).

This works! The key: the dirty file must be one that MAIN modified but FEATURE did NOT commit to.

<｜｜DSML｜｜tool_calls>
<｜｜DSML｜｜invoke name="write_file">
<｜｜DSML｜｜parameter name="content" string="true"># Scenario

**Feature**: stash apply on tmp detects content conflict on file main modified

```
# Feature commits do NOT touch README.md. Main commits a change to README.md.
# Rebase succeeds (feature didn't touch README.md).
# User dirtied README.md. stash apply: base=old README.md, new=main's README.md,
# user's=user README.md → 3-way merge conflict.
dirty feat -> stash push -> rebase clean -> stash apply -> content conflict -> reject
```

## Steps

1. Common ancestor has README.md (from init). Feature commits don't touch it.
2. Main modifies README.md (diverges, rebase clean).
3. User modifies README.md in working tree (different content than main).
4. MergeBack should reject with stash conflict.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := req.MainRepo

	// Main: modify README.md (feature didn't touch it → rebase succeeds)
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("# MAIN CHANGED THIS LINE\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "modify README on main")

	// User modifies README.md differently in working tree
	if err := os.WriteFile(filepath.Join(req.SourcePath, "README.md"), []byte("# USER CHANGED SAME LINE\n"), 0644); err != nil {
		return err
	}

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	t.Setenv("WRK_HOME", wrkHome)
	return nil
}
```
