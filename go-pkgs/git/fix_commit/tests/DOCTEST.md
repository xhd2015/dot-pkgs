# git/fix_commit — rewrite one commit, retarget exact-tip refs

## Version
0.0.2

Classic TDD tree for `github.com/xhd2015/dot-pkgs/go-pkgs/git/fix_commit`.
There is no package yet: `Run` calls `fix_commit.RunCLI` so the suite is
compile-RED until an implementer exists. Do not implement product code here.

`kool git fix-commit` will later be a one-line `RunCLI` dispatch. That glue is
out of this tree. No `label: e2e`. No `os.Stdout` swap, no `Chdir`/`Setenv`.

## DSN (Domain Specific Notion)

### Participants

- **`RunCLI`** — parses `<sha> [OPTIONS]`, plans the rewrite, optionally
  mutates the repo. Writes help/success/dry-run to **stdout**. Writes
  `warning:` lines to **stderr**. Returns fatal errors (`Error: …`); does
  **not** print them (kool `main` prints the returned error).
- **Caller** — this harness. Injects `bytes.Buffer` writers and `-C <dir>`.
  On a returned error, `Run` appends `err.Error()+"\n"` to captured stderr
  so asserts read `resp.Stderr` only.
- **Repo** — local git checkout in `t.TempDir()`, default branch **`master`**.
  Optional local **bare** remotes. Never the network.
- **Git** — binary on PATH. Skip the leaf when missing.

### Behaviors

- **Help** — `-h`/`--help` prints usage (all flags below), exit 0, trailing `\n`.
- **Validate first** — flag parse and “at least one change flag” / mutual
  exclusion run **before** resolving SHA. Missing trailer / empty leftover /
  nothing-to-change hard-error even under `--dry-run`.
- **Rewrite** — `git commit-tree` same tree + same parents. Change only the
  requested author fields. Preserve committer name/email and **both** dates.
  `-m` replaces the full message. `--strip-co-author` is line-based (trim
  including `\r`; line **starts with** `co-authored-by:` case-insensitive,
  colon required); strip all matches; collapse a now-empty trailer block;
  trim trailing whitespace; keep one terminating newline.
- **Refs** — `git update-ref` every local branch whose tip **is** the old SHA
  (works if that branch is checked out in another worktree). Detached `HEAD`
  on the old SHA moves too. Other branches stay. Descendants are **not**
  rewritten; warn and continue. Branch/tag names in output are lexicographic.
- **Tags** — tags that **peel** to the old SHA: capture type + annotated
  payload, `git tag -d`, remote `--delete` only if `ls-remote` still shows
  that tag at the old commit, recreate (lightweight vs annotated), push.
  Default remote `origin`; `--remote` overrides. Missing remote / tag at a
  different remote commit: local retag + `warning:`; skip that remote op.
  Remote exists but has no such tag: local retag + stdout
  `notice: remote <remote> has no tag <name>`; skip remote delete/push.
- **`--push`** — `git push --force-with-lease` each **updated** branch whose
  **upstream** still points at the old SHA. No branch push without `--push`.
- **One pipeline** — discovery, validation, strip parse, ref listing, and
  `ls-remote` run in live **and** dry-run. Side effects (`commit-tree`,
  `update-ref`, `tag -d`, push, retag) skip when `--dry-run`. Dry-run stdout
  lines start with `[dry-run]`; print `would rewrite <oldsha>` (no new SHA).

### Error contract

`RunCLI` returns the fatal string (must start with `Error:`) and does not write
it to stderr. Warnings (`warning:` prefix) go to stderr and do not fail the
command. A missing remote tag is a `notice:` on stdout, not a warning.
This harness appends the returned error to `resp.Stderr`.

### Locked usage

```
Usage: kool git fix-commit <sha> [OPTIONS]

  -m, --message <msg>     replace the full commit message
  --name <name>           replace the author name
  --email <email>         replace the author email
  --strip-co-author       remove Co-authored-by lines from the message;
                          errors if none are present
  --remote <name>         remote for tag delete/push (default: origin)
  --push                  also force-with-lease push updated branches
                          whose upstream still points at the old sha
  --dry-run               print the plan; do not rewrite refs or remotes
  -C, --dir <dir>         repository directory (default: current directory)
  -h, --help              show this help
```

### Locked success / dry-run / errors

```
rewrote <old> -> <new>
  author:  <name> <<email>>
  message: <first line>
  stripped Co-authored-by
  branches:
    master
  tags:
    v0.0.333  delete local+origin, retag, push
  pushed:
    master
```

`stripped Co-authored-by`, `tags:`, and `pushed:` appear only when that action
applied. Tag action verbs: `delete local+<remote>, retag, push` when the remote
tag was deleted; `delete local, retag` when only local retag ran.

```
[dry-run] would rewrite <old>
[dry-run]   author:  <name> <<email>>
[dry-run]   message: <first line>
[dry-run]   stripped Co-authored-by
[dry-run]   move branch master
[dry-run]   delete local tag v0.0.333
[dry-run]   git push origin --delete refs/tags/v0.0.333
[dry-run]   tag v0.0.333 at new commit
[dry-run]   git push origin v0.0.333
```

```
Error: fix-commit requires <sha>
Error: at least one of -m, --email, --name, --strip-co-author is required
Error: --strip-co-author and -m cannot be used together
Error: unknown flag: --nope
Error: not a git repository: <dir>
Error: unknown revision: not-a-real-sha
Error: commit message has no Co-authored-by line
Error: nothing to change
Error:
```

(`Error:` alone = leftover message empty after strip.)

```
warning: commit has descendants; those commits still parent <oldsha>
warning: remote origin tag v0.0.333 points at a different commit; skip remote delete
warning: remote origin not found; skip remote tag operations
```

## Decision Tree

```
fix_commit
├── help/
│   └── show-usage                    --help usage, exit 0
├── validation/                       reject before rewrite plan
│   ├── missing-sha                   no positional
│   ├── missing-fields                no -m/--email/--name/--strip-co-author
│   ├── unknown-flag                  --nope
│   ├── strip-and-message             --strip-co-author + -m
│   ├── not-a-git-repo                -C points at a plain dir
│   └── unknown-revision              sha does not resolve
└── rewrite/                          valid repo + sha + at least one change flag
    ├── message/                      -m is the message source
    │   ├── live/
    │   │   ├── replaced              message only; tip moves; identity/dates/tree/parent kept
    │   │   ├── with-author           -m + --name + --email
    │   │   ├── nothing-to-change     -m already matches
    │   │   ├── branches              every exact-tip branch moves; unrelated stays
    │   │   ├── detached-head         detached HEAD on old sha moves
    │   │   ├── descendants-warn      child still parents old; warning; tip still moves
    │   │   ├── worktree-tip          branch checked out in another worktree still moves
    │   │   ├── push                  --push force-with-lease when upstream at old sha
    │   │   └── tags/
    │   │       ├── lightweight       local delete, remote --delete, retag, push
    │   │       ├── annotated         annotation + tagger preserved
    │   │       ├── remote-points-elsewhere  skip remote delete; local retag
    │   │       ├── no-origin         local retag; warning skip remote
    │   │       ├── remote-absent     origin exists, no such tag; notice
    │   │       └── custom-remote     --remote other; origin tag stays
    │   └── dry-run/
    │       ├── plan                  [dry-run] plan with tag ops; no mutations
    │       └── no-remote-tag         origin exists, no such tag; notice; no remote ops
    ├── author/                       no -m, no strip
    │   ├── name-only
    │   ├── email-only
    │   ├── both
    │   └── nothing-to-change         --name already matches
    └── strip/                        --strip-co-author is the message source
        ├── removes-trailers          multiple + mixed case + CR; collapse blank
        ├── missing                   no matching line
        ├── prose-without-colon       "co-authored by" is not a trailer
        ├── empty-after-strip         message was only the trailer → Error:
        ├── with-name                 strip + --name
        ├── dry-run                   plan shows stripped subject; no mutations
        └── dry-run-missing           missing trailer still fatal under --dry-run
```

### Parameter significance (high → low)

1. **Outcome class** — help / validation / rewrite.
2. **Message source** — `-m` / `--strip-co-author` / author-only.
3. **Ref topology** — branches / detached HEAD / descendants / worktree / `--push`.
4. **Tag + remote state** — lightweight / annotated / elsewhere / missing / `--remote`.
5. **Live vs dry-run**.

## Test Index

| Leaf | Description |
|------|-------------|
| `help/show-usage` | `--help` prints locked usage, exit 0, trailing `\n` |
| `validation/missing-sha` | `Error: fix-commit requires <sha>` |
| `validation/missing-fields` | none of `-m/--email/--name/--strip-co-author` |
| `validation/unknown-flag` | `Error: unknown flag: --nope` |
| `validation/strip-and-message` | `--strip-co-author` and `-m` mutually exclusive |
| `validation/not-a-git-repo` | `-C` is not a git repo |
| `validation/unknown-revision` | sha does not resolve |
| `rewrite/message/live/replaced` | `-m` only; new sha; author/dates/tree/parent kept; `master` moves |
| `rewrite/message/live/with-author` | `-m` + `--name` + `--email`; committer + dates kept |
| `rewrite/message/live/nothing-to-change` | `-m` equals current message → `Error: nothing to change` |
| `rewrite/message/live/branches` | every `--points-at` branch moves; unrelated branch does not |
| `rewrite/message/live/detached-head` | detached HEAD on old sha moves; `master` parked elsewhere stays |
| `rewrite/message/live/descendants-warn` | child still parents old sha; `warning:`; exact-tip branch moves |
| `rewrite/message/live/worktree-tip` | `update-ref` moves a branch checked out in a linked worktree |
| `rewrite/message/live/push` | `--push` force-with-lease for upstream-at-old; no-upstream branch not pushed |
| `rewrite/message/live/tags/lightweight` | delete local+origin, retag, push; remote now new sha |
| `rewrite/message/live/tags/annotated` | annotation and tagger preserved on new commit |
| `rewrite/message/live/tags/remote-points-elsewhere` | skip remote delete with `warning:`; local retag |
| `rewrite/message/live/tags/no-origin` | local retag; `warning:` skip remote |
| `rewrite/message/live/tags/remote-absent` | origin exists, tag not on origin; `notice:` on stdout |
| `rewrite/message/live/tags/custom-remote` | `--remote other`; other moves, origin tag stays |
| `rewrite/message/dry-run/plan` | `[dry-run]` plan including tag remote ops; refs unchanged |
| `rewrite/message/dry-run/no-remote-tag` | `[dry-run]` notice; no remote delete/push; refs unchanged |
| `rewrite/author/name-only` | `--name` only; email/message/dates/tree/parent kept |
| `rewrite/author/email-only` | `--email` only |
| `rewrite/author/both` | `--name` + `--email`; message unchanged |
| `rewrite/author/nothing-to-change` | `--name` already matches → `Error: nothing to change` |
| `rewrite/strip/removes-trailers` | strip all matching lines; leftover subject; `stripped` line |
| `rewrite/strip/missing` | no trailer → error; sha unchanged |
| `rewrite/strip/prose-without-colon` | `co-authored by` is not a match → same missing error |
| `rewrite/strip/empty-after-strip` | message was only the trailer → `Error:` |
| `rewrite/strip/with-name` | strip + `--name` both apply |
| `rewrite/strip/dry-run` | plan shows stripped subject; no mutations |
| `rewrite/strip/dry-run-missing` | missing trailer still fatal under `--dry-run` |

## How to Run

From the `external/dot-pkgs-master-2026-08-14-3` directory:

```sh
doctest vet ./go-pkgs/git/fix_commit/tests
doctest test ./go-pkgs/git/fix_commit/tests
```

From the kool worktree root:

```sh
doctest vet ./external/dot-pkgs-master-2026-08-14-3/go-pkgs/git/fix_commit/tests
doctest test ./external/dot-pkgs-master-2026-08-14-3/go-pkgs/git/fix_commit/tests
```

From the `go-pkgs` module root (`go.mod` lives here):

```sh
doctest vet ./git/fix_commit/tests
doctest test ./git/fix_commit/tests
```

Expected RED until `fix_commit.RunCLI` exists.

```go
import (
	"bytes"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/fix_commit"
)

type Request struct {
	Args []string

	Dir         string
	BareRemote  string
	OtherRemote string
	WorktreeDir string

	OldSHA          string
	ParentSHA       string
	TreeSHA         string
	UnrelatedSHA    string
	ChildSHA        string
	AuthorName      string
	AuthorEmail     string
	CommitterName   string
	CommitterEmail  string
	Message         string
	TagName         string
	TagMessage      string
	TaggerName      string
	TaggerEmail     string
	TaggerUnix      string
	UnrelatedBranch string
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	var stdout, stderr bytes.Buffer
	runErr := fix_commit.RunCLI(req.Args, &stdout, &stderr)
	resp := &Response{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
	if runErr != nil {
		resp.ExitCode = 1
		// Contract: RunCLI returns fatal errors and does not write them.
		// Harness appends err.Error()+"\n" so asserts read resp.Stderr.
		msg := runErr.Error()
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		resp.Stderr += msg
	}
	return resp, nil
}
```
