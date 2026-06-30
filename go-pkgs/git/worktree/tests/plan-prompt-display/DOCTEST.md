# Merge-back plan display — `FormatPlanPrompt` and dry-run command listing

## Version
0.0.2

Doc tests for human-readable merge-back output in
`github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree`. Planned `git -C` commands
and confirmation prompts must use **dynamic target branch labels**, **per-command
`#` comments**, and **`pathfmt.Short` paths in display only** (execution keeps
absolute paths — covered by mvd integration tests).

## DSN (Domain Specific Notion)

### Participants

- **`MergeBack`** — compares source worktree branch to target checkout, builds a
  `MergeBackPlan`, optionally prompts, then runs `PlannedCommand` entries via
  `git -C`.
- **`buildMergeBackPlan`** — selects command list from branch **relation**
  (`ahead`, `diverged`, `same`/`ancestor` when included).
- **`MergeBackPlan`** — carries `Relation`, `Branch`, `TargetLabel`, `Commands`,
  and `NeedsConfirm` (true for `ahead` and `diverged`).
- **`PlannedCommand`** — `{Dir, Args}` for execution; display formatting is
  separate from `String()` used for exec logging today.
- **`FormatPlanPrompt`** — formats the confirmation banner (relation-specific
  question, indented planned commands, `Proceed? [Y/n]:`). Used when
  `NeedsConfirm` and `Confirm` runs before execution.
- **`printDryRun`** — prints planned commands to stdout (no prompt header); used
  when `DryRun` is true for any relation that lists commands.
- **`targetLabel`** — resolves the name shown in questions and comments from the
  branch checked out at the merge **target** (`ReadBranch`); detached target
  falls back to directory basename.
- **`pathfmt.Short`** — shortens `-C` directories and `worktree remove` path
  arguments in **display** strings only.

### Behaviors

**Relation → command shape**

- **`ahead`** — `merge --ff-only <branch>` at target; with `Remove`: `worktree
  remove` + `branch -D` at main repo.
- **`diverged`** — `rebase <target-HEAD>` at source, then `merge --ff-only` at
  target; with `Remove`: same cleanup commands as ahead.
- **`same` / `ancestor` (included)** — no merge/rebase; with `Remove` and
  `DryRun`: only cleanup commands are listed.

**Display sink**

- **Confirm prompt** — `FormatPlanPrompt`: question line uses `TargetLabel`;
  each planned command preceded by a comment line; command lines use shortened
  paths.
- **Dry-run listing** — same comment + shortened command lines as the prompt
  body, without the question or `Proceed?` trailer.

**Target label**

- Target is main repo checkout → label is that checkout's current branch (e.g.
  `master` or `main` from `git init`, never a hardcoded `"main"` when checkout
  is `master`).
- Target is another linked checkout → label is `ReadBranch` of that path.
- Detached target → label is `filepath.Base(targetAbs)`.

## Decision Tree

```
plan-prompt-display
├── relation-ahead/                    [CASE B: ff-merge + optional remove]
│   ├── confirm-prompt-with-remove/    FormatPlanPrompt via Confirm abort
│   ├── dry-run-with-remove/           printDryRun stdout
│   └── merge-uses-branch/             attached branch name in merge (not SHA)
├── relation-detached-head/            [detached HEAD ahead of target]
│   └── dry-run/                       merge --ff-only <commit-sha>
├── relation-diverged/                 [CASE C: rebase + ff-merge + remove]
│   ├── confirm-prompt-with-remove/
│   └── dry-run-with-remove/
├── relation-included/                 [ancestor/same; commands only when DryRun]
│   └── dry-run-remove-only/
└── target-label/                      [TargetLabel resolution in prompt]
    ├── explicit-master/               git init -b master → not "main"
    └── separate-target-checkout/      TargetPath = other worktree branch name
```

## Test Index

| Leaf | Sink | Description |
|------|------|-------------|
| `relation-ahead/confirm-prompt-with-remove` | Prompt | Ahead + Remove: comments, Short paths, dynamic target in question |
| `relation-ahead/dry-run-with-remove` | Dry-run | Same command display as prompt body, no Proceed line |
| `relation-ahead/merge-uses-branch` | Dry-run | Attached worktree uses branch name in merge, not commit hash |
| `relation-detached-head/dry-run` | Dry-run | Detached HEAD uses commit SHA; not falsely already-included |
| `relation-diverged/confirm-prompt-with-remove` | Prompt | Rebase comment `# <branch>: rebase onto <target>` + full command set |
| `relation-diverged/dry-run-with-remove` | Dry-run | Diverged command list formatting |
| `relation-included/dry-run-remove-only` | Dry-run | Only `# worktree: remove` and `# worktree branch: drop` |
| `target-label/explicit-master` | Prompt | `merge into master?` and `# master: fast forward` |
| `target-label/separate-target-checkout` | Prompt | Question uses target worktree branch, not main default |

## Related integration tests (mvd)

Update leaves under `go-pkgs/cmd/mvd/tests/` (same display contract end-to-end):

- `mode-worktree-back-enhanced/branch-ahead/prompt-shows-commands`
- `mode-worktree-back-enhanced/branches-diverged/prompt-shows-commands`
- `mode-dry-run/dry-run-back-worktree`

## How to Run

```sh
doctest vet ./go-pkgs/git/worktree/tests/plan-prompt-display
doctest test ./go-pkgs/git/worktree/tests/plan-prompt-display
```

```go
import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

type Request struct {
	WorkRoot   string
	SourcePath string
	TargetPath string // empty → main repo checkout
	DryRun     bool
	// CapturePrompt: when true, Run records FormatPlanPrompt from Confirm (no execution).
	CapturePrompt bool
	// DefaultBranch: non-empty passed to git init (-b), e.g. "master".
	DefaultBranch string
}

type Response struct {
	Output      string
	TargetLabel string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var promptBuf strings.Builder
	var stdout bytes.Buffer
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	os.Stdout = w

	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&stdout, r)
		close(done)
	}()

	var capturedLabel string
	opts := worktree.MergeBackOptions{
		SourcePath: req.SourcePath,
		TargetPath: req.TargetPath,
		DryRun:     req.DryRun,
		Remove:     true,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			capturedLabel = plan.TargetLabel
			if req.CapturePrompt {
				promptBuf.WriteString(worktree.FormatPlanPrompt(plan))
			}
			return false, nil
		},
	}
	result, runErr := worktree.MergeBack(opts)

	w.Close()
	os.Stdout = oldStdout
	<-done

	if runErr != nil {
		return nil, runErr
	}
	if result == nil {
		return nil, nil
	}

	out := stdout.String()
	if req.CapturePrompt {
		out = promptBuf.String()
	}
	return &Response{Output: out, TargetLabel: capturedLabel}, nil
}
```