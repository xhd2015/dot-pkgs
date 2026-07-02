# wrk Test Cases

## Version
0.0.2

Decision tree covering the `wrk` CLI: no-args worktree creation, optional `wrk <dir>`
first positional, `wrk <dir> <target-dir>` second positional spawn-location override,
`wrk --dep` external dependency worktrees, `wrk --done` merge-back
(including external cascade), `wrk --list`, and `wrk --status`.

# DSN (Domain Specific Notion)

- **wrk CLI** — standalone binary; invocation form `wrk [dir] [flags...]`; first non-flag positional argument is optional `<dir>` — when present, effective cwd is the resolved absolute path of `<dir>` (process cwd unchanged); when absent, effective cwd is process cwd; no-args invocation creates a git worktree from the effective checkout and prints the target path on stdout; `wrk --done` merges the linked worktree branch back into main and removes the worktree + branch (`worktree.MergeBack` with `Remove: true`); `wrk --list` runs `git -C <effective-cwd> worktree list` and prints stdout unchanged; `wrk --status` resolves the effective cwd's git toplevel, scans it with `scan_repo.Scan`, and prints one status block per discovered git directory; missing `<dir>` → `wrk: <path> does not exist`.
- **WRK_HOME** — storage root env var (default `~/.wrk`); tests isolate per run at `{WorkRoot}/.wrk`.
- **WRK_DATE** — optional env var (`YYYY-MM-DD`) overriding the run date for deterministic naming; all tests set `WRK_DATE=2026-06-30`.
- **Work root** — temp directory holding source repos and move targets.
- **Naming** — worktree path `{WRK_HOME}/worktrees/{basename}-{token}-{YYYY-MM-DD}[-N]`; branch `{base-branch}-{YYYY-MM-DD}[-N]`; `N` starts at 1 on collision (path exists or branch ref exists). No unsuffixed names without date.
- **token** — `sanitize(base-branch)` for normal branches (`/` → `-` in path only); for detached HEAD, 7-char short commit hash from `git rev-parse --short=7 HEAD` (not literal `HEAD`).
- **Git source** — cwd must be inside a git checkout (main repo, linked worktree, or nested subdirectory); basename resolves from the checkout root when cwd is a linked worktree or nested subpath.
- **wrk --dep** — spawns a dependency worktree under `{consumerTop}/external/` as a worktree of the DEP repo (`git -C <depMain> worktree add`), so it is registered under `<depMain>/.git/worktrees/` (NOT the consumer's); the dep already holds its own objects, so no remote/fetch into the consumer is needed. Appends `/external` to `.gitignore` when missing, runs `gotool.Replace` + `gotool.Tidy`, prints external worktree abs path on stdout. Naming: `{basename}-{token}-{WRK_DATE}[-N]` (same rules as create; basename from dep main repo); the branch lives in the dep repo, so the `-N` collision check runs against `depMain` (path under `consumerTop/external/` + branch in dep repo).
- **wrk --done** — resolves checkout root via `ShowToplevel(cwd)`; requires a linked worktree (not main repo); clean worktree; implicit `--rm`. **Cascade**: merge-back each linked worktree under `{toplevel}/external/*` first — these are dep-repo worktrees (registered under `<depMain>/.git/worktrees/`), so `MergeBack` resolves their main repo from the worktree's `.git` gitdir (the dep main) and merges the dep branch back into the dep repo (the branch shares the dep's history, so merge-base resolves); this ensures dep work committed on an external worktree is merged back before removal. Relation to dep main: already-included → remove only; ahead/diverged → prompt (`--confirm-from-stdin`), non-interactive ahead/diverged falls back to force-removal. **Guard**: scan **every** Go module under the checkout (`gotool/mod/scan.Scan`) — main + all sub-modules — and classify each filesystem/local `replace` (`./`, `../`, or absolute path without version) by resolving its target relative to the module's `go.mod` dir: **intra-repo** = target dir exists AND `git -C <target> rev-parse --show-toplevel` equals the consumer's toplevel (a `../../`/`./sub` nested-module reference back into the same repo); **extra-repo** = everything else (`./external/foo` dep worktree, non-existent target, absolute/sibling outside). The guard names the offending `<top>/<m.Dir>/go.mod` file and each `replace <Old> => <New>` directive in its message. Default (no flag): intra-repo → **WARN to stderr and proceed** (exit 0, merge-back runs); extra-repo → **error, block**. `--no-in-module-replace` (opt-in, valid only with `--done`) → **all** local replaces block (fully-strict). A checkout with no `go.mod` at all yields zero modules → guard is a no-op → `--done` proceeds (and the linked-worktree check inside `MergeBack` still runs for a main-repo cwd, producing `not a linked worktree`). Branch relation to main: already-included → remove only; ahead/diverged → prompt then merge/rebase.
- **wrk --list** — runs `git -C <cwd> worktree list`; prints stdout unchanged; cwd must be inside a git work tree (main repo, linked worktree, or nested subpath). Mutually exclusive with no-args create and `--done`.
- **wrk --status** — standalone reporting mode; cwd must be inside a git work tree. Resolves the effective cwd's checkout root with git toplevel discovery, calls `scan_repo.Scan(context.Background(), scan_repo.Options{Roots: []string{Root}})`, and prints every discovered git directory in scan path order. Each block includes `Dir` relative to the initial toplevel (`.` for the root), current branch, short commit hash plus subject, and `Status` as either `clean` or `dirty (<added> added, <changed> changed, <renamed> renamed, <deleted> deleted)`. Mutually exclusive with `--done`, `--list`, `--dep`, `--all-deps`, create target arguments, and other modes.
- **--confirm-from-stdin** — when set with piped `StdinInput`, reads Y/n from stdin for merge-back confirmation (required for non-TTY ahead/diverged cases).
- **--no-in-module-replace** — bool flag (no value); valid ONLY with `--done`. Restores the fully-strict local-replace guard: every filesystem/local `replace` (intra-repo or extra-repo) blocks `--done`. Without it (default), intra-repo replaces — whose target dir exists and shares the consumer's `git rev-parse --show-toplevel` (`../../`/`./sub` nested-module reference) — only WARN and `--done` proceeds; extra-repo replaces (`./external/foo` dep worktree, non-existent/absolute/sibling) still block. Bare `wrk --no-in-module-replace`, or with any other mode (`--dep`/`--list`/no-args create/`--all-deps`) → non-zero exit, stderr `wrk: --no-in-module-replace is only valid with --done`.
- **Request.Args** — CLI arguments passed to `wrk` after optional `<dir>` (empty → no-args create; `["--dep", depPath]` for dep tests; `["--done"]` or `["--done", "--confirm-from-stdin"]` for done tests; `["--list"]` for list tests).
- **Request.TargetDir** — when set, prepended as the first positional argument to `wrk` (`wrk <dir> ...`); used by `dir-arg/` tests to run from `WorkRoot` while targeting a repo elsewhere.
- **Request.SpawnDir** — optional second positional `<target-dir>` (`wrk <dir> <target-dir>`); appended after `TargetDir` when set. Overrides the worktree spawn location: missing target with existing parent → spawn exactly at `<target-dir>` (no naming suffix on the path); existing target dir → spawn a default-named sub-dir under it; missing parent → error. Resolved relative to the process (shell) cwd, not `<dir>`. Create-only — errors with `wrk: unexpected arguments` when combined with `--list`/`--done`/`--dep`. `WRK_HOME` is ignored when set.
- **external/** — dependency worktrees live at `{consumerTop}/external/{basename}-{token}-{WRK_DATE}[-N]`; not under `WRK_HOME`. They are linked worktrees of the DEP repo (registered under `<depMain>/.git/worktrees/`), not the consumer — the consumer only hosts the working tree on disk.
- **gitWorktreeList** — helper capturing raw `git worktree list` stdout from a directory for list-test comparison.
- **Request.StdinInput** — when non-empty, piped to wrk stdin before wait (mvd merge-back pattern).
- **wrk --all-deps** — automates `--dep` for every dependency that has a local git repo: reads the consumer's `go.mod` (`resolve.GetModuleInfo`) for the required-module set, existing local `Replace` set, and the consumer's own module path; scans scan roots (`scan_repo.Scan`, `RepoTypeMain` only, sorted by path) for git repos whose module path matches a required module; skips self, modules not required, already-replaced modules (tolerated, not errored — unlike `--done`), and already-seen modules; for each match links an external worktree under `{consumerTop}/external/` via the shared `linkExternalDep` core and records a `replace`; runs `go mod tidy` once at the end (skipped when zero deps linked). Stdout: one line per linked dep in scan (path-sorted) order `wrk <module-path> at ./external/<name>` (path relative to `consumerTop`), then a final summary `wrk <N> deps`. Zero deps → single line `wrk 0 deps`, exit 0, no tidy, no `external/` created.
- **WRK_SCAN_ROOT** — env var overriding the scan root for `wrk --all-deps` when `--scan-root` is not set (expanded via `pathfmt.Expand` / home expansion like `scan_repo`); precedence: `--scan-root` flag > `WRK_SCAN_ROOT` env > home dir (`~`). The resolved `WRK_HOME` directory is added to ignore dirs so the worktree store is not walked.
- **--scan-root** — value-consuming flag for `wrk --all-deps` overriding the scan root; like `--dep` it consumes its following value argument and is not mistaken for the optional first positional `<dir>`.
- **--all-deps mutual exclusion** — `--all-deps` is mutually exclusive with `--dep`, `--done`, `--list`, and no-args create; `--all-deps --dep <x>` → non-zero exit, stderr mentions "mutually exclusive"; no positional args allowed.
- **wrk --all-deps --dry-run** — runs the full read-only discovery/planning of `--all-deps` but writes nothing. Same cwd/git/go.mod validation, same `required`/`alreadyReplaced`/`consumerModule` sets, same scan-root resolution (`--scan-root` > `WRK_SCAN_ROOT` > home), same `scan_repo.Scan` + `modscan.Scan`, same self / not-required / already-replaced / seen skips, and the SAME external-path naming + collision logic as the real run — but it does NOT `MkdirAll(external/)`, does NOT `ensureGitignoreExternal`, does NOT `createExternalWorktree` (no `git worktree add`/branch/remote/fetch), does NOT `GoModEditReplace`, and does NOT `GoModTidy`. stdout: one line per planned module in scan order `would: wrk <module-path> at ./external/<name>[/<subdir>]`, then a final `would: wrk <N> deps` (zero → single `would: wrk 0 deps`). Core guarantee: after a dry run `{consumerTop}/external/` does NOT exist, consumer `go.mod` is unchanged (no new replaces), and `.gitignore` is unchanged (no `/external` line). Errors that occur during planning (non-git cwd, unreadable go.mod) still surface as errors — the process "actually runs".
- **--dry-run** — bool flag (no value); valid ONLY with `--all-deps` (and `--scan-root`). Bare `wrk --dry-run`, or `--dry-run` with any other mode (`--dep`/`--done`/`--list`/no-args create) → non-zero exit, stderr `wrk: --dry-run is only valid with --all-deps`. It does NOT relax `--all-deps`'s mutual exclusion with `--dep`/`--done`/`--list` — `--dry-run --all-deps --dep <x>` still errors as mutually exclusive (the `--all-deps` mutual-exclusion check runs first). `extractDir` in `cmd/wrk/main.go` needs no change (its `strings.HasPrefix(arg, "-")` branch already treats `--dry-run` as a flag).
- **wrk --task <desc>** — flag valid only in create mode (no `--done`/`--list`/`--dep`/`--all-deps`). Derives a sanitized slug from `<desc>` (lowercase, non-letter-digit → `-`, collapse, trim, truncate 64 runes). Appends slug after the date in both dir and branch names: `{basename}-{token}-{date}-{slug}[-N]` for dir, `{branchBase}-{date}-{slug}[-N]` for branch. Empty `<desc>` or slug → non-zero exit. No metadata file stored — the slug is embedded in the name.
- **wrk --set-task <desc>** — flag valid alone (mutually exclusive with all other flags and positional args). Must be run inside a linked worktree. Parses the current branch name to extract `branchBase` and `date` (branch must match `{branchBase}-{YYYY-MM-DD}[-slug][-N]`; non-wrk worktrees → error). Computes new dir and branch names with the new slug. If slug is unchanged → no-op. Before `git worktree move`, checks stdout: TTY → warns (old→new path + branch) and prompts `Proceed? [Y/n]`; confirmation executes the move. Non-TTY → non-zero exit `wrk: --set-task requires a terminal (tty)`. When run with `WRK_SET_TASK_CONFIRM=1` env → auto-confirms without TTY (test escape hatch).
- **Request.TaskDesc** — when set, used by test assertions to compute expected dir/branch names with task slug.
- **Request.SetTaskDesc** — when set, tests pass `--set-task <desc>` to wrk; test assertions verify rename side effects.
- **Request.SetTaskEnv** — when set, appended to wrk's environment (e.g., `WRK_SET_TASK_CONFIRM=1` to auto-confirm rename in tests).

## Tree Overview

```
wrk tests
├── create-worktree/              # cwd is a git checkout (success path)
│   ├── main-checkout/            # cwd is the main repo checkout
│   │   ├── basic-create/         # first wrk from main
│   │   ├── sequence-increment/   # second wrk increments -N suffix
│   │   ├── branch-collision/     # branch ref blocks date-suffixed name
│   │   └── slash-branch/         # branch with / sanitized in path token
│   ├── from-linked-worktree/     # cwd is linked worktree; basename from main repo
│   ├── from-git-subpath/         # cwd is nested subdir inside checkout; basename from repo root
│   └── detached-head/            # cwd on detached HEAD → 7-char hash token
├── dep/                          # wrk --dep external dependency worktree
│   ├── basic/                    # require + --dep → external wt, replace, tidy, gitignore
│   ├── gitignore-already/        # /external already in .gitignore → no duplicate
│   ├── not-a-dependency/         # dep not in go.mod → error
│   ├── not-git-repo/             # dep path not git → error
│   ├── not-go-module/            # dep git without go.mod → error
│   ├── sub-module/               # dep root has no go.mod; module in subdir consumer requires → external wt + replace => <external>/sub
│   ├── external-wt-owned-by-dep-repo/  # external wt .git gitdir points into dep main, not consumer
│   └── external-wt-from-linked-consumer/ # --dep from inside a linked consumer wt → owned by dep main, not consumer main
├── all-deps/                     # wrk --all-deps scan consumer go.mod + local repos → link each matched dep
│   ├── basic/                    # dep1+dep2 both present → both linked, 2 deps
│   ├── partial-local/            # only dep1 present → dep1 linked, dep2 not replaced, 1 deps
│   ├── already-replaced/         # dep1 pre-replaced → skipped; dep2 linked, 1 deps
│   ├── no-local-deps/            # scan-root empty → 0 deps, no replaces, no external/
│   ├── self-skip/                # consumer inside scan-root alongside dep1 → dep1 only, self skipped
│   ├── mutually-exclusive/       # wrk --all-deps --dep <x> → non-zero, mutually exclusive
│   ├── not-git-cwd/              # cwd not a git repo → non-zero, is not a git repository
│   ├── nested-submodule/         # required module nested in a larger repo → linked via sub-module discovery
│   ├── multi-module-same-repo/   # two required sub-modules in one repo → one shared worktree, two replaces
│   └── dry-run/                  # wrk --all-deps --dry-run: plan only, write nothing
│       ├── basic/                # dep1+dep2 present → would: lines for both, no side effects
│       ├── no-matches/           # scan-root empty → would: wrk 0 deps, no side effects
│       └── without-all-deps/     # wrk --dry-run (no --all-deps) → non-zero, only valid with --all-deps
├── done/                         # wrk --done merge-back --rm from linked worktree
│   ├── already-included/         # branch already merged into main; remove only
│   ├── ahead-confirm/            # ahead + --confirm-from-stdin + Enter
│   ├── ahead-decline/            # ahead + --confirm-from-stdin + decline
│   ├── ahead-non-tty/            # ahead without confirm flag (non-interactive)
│   ├── dirty/                    # uncommitted changes in worktree
│   ├── not-linked/               # cwd is main repo (not linked worktree)
│   ├── from-subpath/             # cwd nested inside linked wt; uses checkout root
│   ├── external-cascade/         # cascade removes external/* wt; guard blocks parent (names go.mod + directive)
│   ├── cascade-merge-base/       # cascade must remove dep wt, not crash "failed to find merge base" (dep branch shares no history with consumer main)
│   ├── cascade-dep-merge-back/   # ahead dep wt → cascade ff-merges dep branch into dep repo, then removes wt (merge-back, not discard)
│   ├── local-replace-blocks/     # extra-repo fs replace (non-existent ./external/foo) → guard blocks + names go.mod + directive
│   ├── intra-replace-warns/      # intra-repo fs replace (./submod, same toplevel) → WARN + proceed (default, exit 0)
│   ├── intra-replace-strict-blocks/ # intra-repo replace + --no-in-module-replace → block + names go.mod + directive
│   ├── no-in-module-replace-without-done/ # --no-in-module-replace without --done → error
│   ├── no-go-mod/                # linked wt whose checkout has no go.mod → --done merge-back succeeds (guard is no-op)
│   ├── not-linked-no-go-mod/     # main repo without go.mod → "not a linked worktree" (go.mod check must not mask it)
│   └── sub-module-replace-blocks/ # main go.mod clean but sub/go.mod has local replace → guard blocks + names go.mod + directive
├── list/                         # wrk --list (git worktree list wrapper)
│   ├── main-only/                # single main checkout, no linked worktrees
│   ├── with-linked/              # main + one linked worktree
│   ├── from-subpath/             # cwd nested inside main repo
│   └── non-git/                  # cwd is not a git repo (error)
├── status/                       # wrk --status status-block display
│   ├── valid-git-cwd/            # cwd resolves to a git checkout
│   │   ├── root-clean/           # root checkout shown as "." and clean
│   │   ├── subdir-clean/         # nested cwd still reports root-relative "."
│   │   ├── multiple-git-dirs/    # root + nested independent repo blocks
│   │   └── dirty-counts/         # added/changed/renamed/deleted counts
│   ├── invalid-git-cwd/
│   │   └── non-git/              # cwd is not a git repo (error)
│   └── invalid-mode/
│       └── with-list/            # --status with --list is mutually exclusive
├── non-git-cwd/                  # cwd is not a git repo (error, no-args create)
├── dir-arg/                      # wrk <dir> optional first positional
│   ├── create/
│   │   └── basic/                # wrk <repoDir> from WorkRoot creates worktree
│   ├── list/
│   │   └── from-dir/             # wrk <repoDir> --list from WorkRoot
│   └── missing-dir/              # wrk <nonexistent> → does not exist
└── target-dir/                   # wrk <dir> <target-dir> custom spawn location
    ├── target-missing/           # <target-dir> does not exist
    │   ├── parent-exists/        # spawn exactly at <target-dir> (case 1)
    │   └── parent-missing/       # parent missing → error (case 3)
    ├── target-exists/            # <target-dir> exists
    │   ├── basic-subdir/         # spawn default-named sub-dir under it (case 2)
    │   ├── collision-suffix/     # sub-dir name collides → -N suffix (case 2)
    │   └── target-is-file/       # target is a file → error (edge)
    ├── relative-path/            # relative <target-dir> resolved vs shell cwd
    └── with-other-mode/          # target-dir + other mode → error
        ├── with-list/            # wrk <dir> <target-dir> --list
        └── with-dep/             # wrk <dir> <target-dir> --dep <dep>
    └── task/                          # wrk --task and wrk --set-task
        ├── spawn/                     # --task when creating worktree
        │   ├── basic/                 # wrk --task "fix login bug" → slug in name
        │   ├── special-chars/         # capitals, symbols → sanitized slug
        │   ├── long-task/             # >64 runes → truncated
        │   ├── empty-task/            # --task "" → error
        │   ├── empty-slug/            # --task "!!!" → error (slug empty)
        │   ├── with-done/             # --task + --done → mutually exclusive
        │   ├── sequence/              # two --task "same" calls → -N suffix
        │   ├── branch-collision/      # pre-existing branch blocks → suffix
        │   └── target-dir/            # wrk <dir> <target> --task → branch has slug
        └── set-task/                  # --set-task inside linked worktree
            ├── non-tty/               # non-TTY → "requires terminal" error
            ├── empty-desc/            # --set-task "" → error
            ├── empty-slug/            # --set-task "!!!" → error
            ├── not-linked/            # from main repo → error
            └── not-wrk-worktree/      # custom branch → cannot parse → error
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | create-worktree/main-checkout/basic-create | Fresh git repo on `main`, first `wrk` |
| 2 | create-worktree/main-checkout/sequence-increment | Run `wrk` twice from same repo/branch |
| 3 | create-worktree/main-checkout/branch-collision | Pre-create branch `main-{date}`, no worktree at date path |
| 4 | non-git-cwd | cwd is not a git repository |
| 5 | create-worktree/from-linked-worktree | cwd is existing linked worktree |
| 6 | create-worktree/detached-head | cwd on detached HEAD uses 7-char hash token |
| 7 | create-worktree/main-checkout/slash-branch | Branch `feature/foo` |
| 8 | create-worktree/from-git-subpath | cwd is nested subdir inside checkout; basename from repo root |
| 9 | done/already-included | wt branch merged into main; `--done` removes wt + branch |
| 10 | done/ahead-confirm | wt ahead; `--done --confirm-from-stdin` + `\n` → ff-merge + remove |
| 11 | done/ahead-decline | wt ahead; `--confirm-from-stdin` + `n\n` → aborted, wt remains |
| 12 | done/ahead-non-tty | wt ahead; no confirm flag → non-zero (cannot prompt) |
| 13 | done/dirty | uncommitted file in wt → non-zero |
| 14 | done/not-linked | cwd is main repo → `not a linked worktree` |
| 15 | done/from-subpath | cwd is subdir inside linked wt; `--done` uses checkout root |
| 16 | list/main-only | single main repo; stdout matches `git worktree list` |
| 17 | list/with-linked | main + linked worktree; stdout lists both paths |
| 18 | list/from-subpath | cwd nested in main repo; stdout same as from repo root |
| 19 | list/non-git | cwd is not a git repository |
| 20 | dep/basic | Consumer requires dep; `--dep` creates external wt, replace, tidy, gitignore |
| 21 | dep/gitignore-already | `/external` already present → no duplicate line |
| 22 | dep/not-a-dependency | Dep not in go.mod → error |
| 23 | dep/not-git-repo | Path not git → error |
| 24 | dep/not-go-module | Git dir without go.mod → error |
| 25 | dep/sub-module | Dep root has no go.mod; module in subdir that consumer requires → external wt + replace => `<external>/sub` |
| 26 | dep/external-wt-owned-by-dep-repo | External dep worktree's `.git` gitdir resolves to the DEP main repo, not the consumer; dep owns the worktree |
| 27 | dep/external-wt-from-linked-consumer | `--dep` from inside a linked consumer worktree → external wt owned by dep main repo, not consumer main |
| 28 | all-deps/basic | dep1+dep2 both in scan-root → both linked, `wrk 2 deps` |
| 29 | all-deps/partial-local | only dep1 in scan-root → dep1 linked, dep2 not replaced, `wrk 1 deps` |
| 30 | all-deps/already-replaced | dep1 pre-replaced `=> ./external/preexisting` → skipped; dep2 linked, `wrk 1 deps` |
| 31 | all-deps/no-local-deps | scan-root empty → `wrk 0 deps`, no replaces, no `external/` |
| 32 | all-deps/self-skip | consumer inside scan-root alongside dep1 → dep1 linked, self skipped, `wrk 1 deps` |
| 33 | all-deps/mutually-exclusive | `wrk --all-deps --dep <x>` → non-zero, mutually exclusive |
| 34 | all-deps/not-git-cwd | cwd not a git repo → non-zero, is not a git repository |
| 35 | all-deps/nested-submodule | required module nested in a larger repo → linked via sub-module discovery, replace at sub-dir, `wrk 1 deps` |
| 36 | all-deps/multi-module-same-repo | two required sub-modules in one repo → ONE shared worktree, two replaces, `wrk 2 deps` |
| 37 | all-deps/dry-run/basic | dep1+dep2 in scan-root → `would:` lines for both, `would: wrk 2 deps`, NO `external/`, NO replaces, NO `.gitignore` change |
| 38 | all-deps/dry-run/no-matches | scan-root empty → `would: wrk 0 deps`, NO `external/`, NO replaces, NO `.gitignore` change |
| 39 | all-deps/dry-run/without-all-deps | `wrk --dry-run` (no `--all-deps`) → non-zero, stderr `--dry-run is only valid with --all-deps` |
| 40 | done/external-cascade | `--done` cascades to `external/*` dep wt first; parent errors on local replace (names go.mod + directive) |
| 41 | done/local-replace-blocks | extra-repo `replace => ./external/foo` (non-existent) blocks `--done` at guard (names go.mod + directive) |
| 42 | dir-arg/create/basic | `wrk <repoDir>` from WorkRoot creates worktree |
| 43 | dir-arg/list/from-dir | `wrk <repoDir> --list` matches `git worktree list` |
| 44 | dir-arg/missing-dir | `wrk <nonexistent>` → does not exist |
| 45 | target-dir/target-missing/parent-exists | `wrk <dir> <target>` spawns exactly at `<target>` (parent exists) |
| 46 | target-dir/target-missing/parent-missing | `wrk <dir> <target>` parent missing → error |
| 47 | target-dir/target-exists/basic-subdir | `wrk <dir> <target>` existing dir → default-named sub-dir |
| 48 | target-dir/target-exists/collision-suffix | existing dir + colliding sub-dir → `-N` suffix |
| 49 | target-dir/target-exists/target-is-file | `<target>` is a file → error |
| 50 | target-dir/relative-path | relative `<target>` resolved against shell cwd |
| 51 | target-dir/with-other-mode/with-list | `wrk <dir> <target> --list` → unexpected arguments |
| 52 | target-dir/with-other-mode/with-dep | `wrk <dir> <target> --dep <dep>` → unexpected arguments |
| 53 | done/no-go-mod | linked wt whose checkout has no go.mod; `--done` merge-back succeeds (guard no-op) |
| 54 | done/not-linked-no-go-mod | main repo without go.mod; `--done` → `not a linked worktree` (not `no go.mod found`) |
| 55 | done/sub-module-replace-blocks | main go.mod clean but `sub/go.mod` has local replace → guard blocks `--done` (names go.mod + directive) |
| 56 | done/cascade-merge-base | cascade must remove dep wt, not crash `failed to find merge base` (dep branch vs consumer main share no history) |
| 57 | done/cascade-dep-merge-back | ahead dep wt + `--confirm-from-stdin` → cascade ff-merges dep branch into dep repo, removes wt (merge-back, not discard) |
| 58 | done/intra-replace-warns | intra-repo `replace example.com/foo => ./submod` (existing, same toplevel) → WARN, exit 0, merge-back proceeds |
| 59 | done/intra-replace-strict-blocks | intra-repo replace + `--no-in-module-replace` → block, names go.mod + directive |
| 60 | done/no-in-module-replace-without-done | `wrk --list --no-in-module-replace` → non-zero, `--no-in-module-replace is only valid with --done` |
| 61 | task/spawn/basic | `wrk --task "fix login bug"` → dir/branch include `-fix-login-bug` |
| 62 | task/spawn/special-chars | Task with capitals, symbols, unicode → sanitized slug |
| 63 | task/spawn/long-task | >64 runes → truncated to 64 |
| 64 | task/spawn/empty-task | `--task ""` → error |
| 65 | task/spawn/empty-slug | `--task "!!!"` → error (slug empty after sanitization) |
| 66 | task/spawn/with-done | `--task` + `--done` → mutually exclusive |
| 67 | task/spawn/sequence | Two `wrk --task "same"` calls → `-N` suffix on second |
| 68 | task/spawn/branch-collision | Pre-existing branch with task-slug name → suffix increment |
| 69 | task/spawn/target-dir | `wrk <dir> <target> --task "desc"` → branch has slug, dir is user-specified |
| 70 | task/set-task/non-tty | `--set-task` in non-TTY → error "requires terminal" |
| 71 | task/set-task/empty-desc | `--set-task ""` → error |
| 72 | task/set-task/empty-slug | `--set-task "!!!"` → error |
| 73 | task/set-task/not-linked | `--set-task` from main repo → error |
| 74 | task/set-task/not-wrk-worktree | `--set-task` on custom-branch worktree → cannot parse → error |
| 75 | status/valid-git-cwd/root-clean | `wrk --status` from repo root shows `Dir: .` and clean status |
| 76 | status/valid-git-cwd/subdir-clean | `wrk --status` from nested subdir still shows `Dir: .` |
| 77 | status/valid-git-cwd/multiple-git-dirs | root + nested independent git repo produce two status blocks |
| 78 | status/valid-git-cwd/dirty-counts | status counts one added, changed, renamed, and deleted entry |
| 79 | status/invalid-git-cwd/non-git | `wrk --status` outside git fails with `is not a git repository` |
| 80 | status/invalid-mode/with-list | `wrk --status --list` fails as mutually exclusive |

## How to Run

```sh
# Verify tree structure (no test execution)
doctest vet ./tests

# Run all tests (expect RED until wrk is implemented)
doctest test ./tests

# Run a specific leaf
doctest test ./tests/create-worktree/main-checkout/basic-create

# Run a done leaf
doctest test ./tests/done/ahead-confirm

# Run a list leaf
doctest test ./tests/list/main-only

# Run status leaves
doctest test ./tests/status
doctest test ./tests/status/valid-git-cwd/dirty-counts

# Run a dep leaf
doctest test ./tests/dep/basic

# Run an all-deps leaf
doctest test ./tests/all-deps/basic

# Run a dry-run leaf
doctest test ./tests/all-deps/dry-run/basic

# Run a done cascade leaf
doctest test ./tests/done/external-cascade

# Run a local-replace guard leaf
doctest test ./tests/done/local-replace-blocks
doctest test ./tests/done/sub-module-replace-blocks

# Run an intra-replace (lenient/strict) leaf
doctest test ./tests/done/intra-replace-warns
doctest test ./tests/done/intra-replace-strict-blocks

# Run the --no-in-module-replace validation leaf
doctest test ./tests/done/no-in-module-replace-without-done

# Run a dir-arg leaf
doctest test ./tests/dir-arg/create/basic
doctest test ./tests/dir-arg/list/from-dir
doctest test ./tests/dir-arg/missing-dir

# Run a target-dir leaf
doctest vet ./tests/target-dir
doctest test ./tests/target-dir/target-missing/parent-exists
doctest test ./tests/target-dir/target-exists/collision-suffix
doctest test ./tests/target-dir/relative-path

# Run a task spawn leaf
doctest test ./tests/task/spawn/basic
doctest test ./tests/task/spawn/empty-task
doctest test ./tests/task/spawn/sequence

# Run a task set-task leaf (non-TTY, expects error)
doctest test ./tests/task/set-task/non-tty
doctest test ./tests/task/set-task/empty-desc
doctest test ./tests/task/set-task/not-linked
```

```go
import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"testing"
)

type Request struct {
	WorkRoot   string
	WrkHome    string
	RepoDir    string   // process cwd when running wrk
	TargetDir  string   // optional first positional <dir>; prepended to Args when set
	SpawnDir   string   // optional second positional <target-dir>; appended after TargetDir when set
	HashToken  string   // detached-head: 7-char short commit hash
	Args       []string // CLI args after <dir>; empty → no-args create
	StdinInput string   // piped to stdin when set
	MainRepo      string // done tests: main checkout path
	WtDir         string // done/dep tests: linked worktree path
	WtBranch      string // done tests: worktree branch name
	DepPath       string // dep tests: path-to-repo argument
	ConsumerTop   string // dep tests: consumer git toplevel
	ExternalWtDir string // dep/done tests: external worktree path
	DepModulePath string // dep tests: module path from dep go.mod
	TaskDesc      string // task tests: task description passed to --task
	SetTaskDesc   string // task tests: new task description for --set-task
	SetTaskEnv    string // task tests: extra env vars for --set-task (e.g., WRK_SET_TASK_CONFIRM=1)
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getWrkBin(t)

	var args []string
	if req.TargetDir != "" {
		args = append(args, req.TargetDir)
	}
	if req.SpawnDir != "" {
		args = append(args, req.SpawnDir)
	}
	if req.TaskDesc != "" {
		args = append(args, "--task", req.TaskDesc)
	}
	if req.SetTaskDesc != "" {
		args = append(args, "--set-task", req.SetTaskDesc)
	}
	args = append(args, req.Args...)

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	env := append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)
	if req.SetTaskEnv != "" {
		env = append(env, req.SetTaskEnv)
	}
	cmd.Env = env

	if req.StdinInput != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		if _, err := io.WriteString(stdin, req.StdinInput); err != nil {
			return nil, err
		}
		stdin.Close()
		err = cmd.Wait()
		exitCode := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				exitCode = ee.ExitCode()
			} else {
				return nil, err
			}
		}
		return &Response{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: exitCode,
		}, nil
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}

	return &Response{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}, nil
}
```
