# wrk Test Cases

## Version
0.0.2

Decision tree covering the `wrk` CLI: no-args worktree creation, optional `wrk <dir>`
first positional, `wrk <dir> <target-dir>` second positional spawn-location override,
`wrk --dep` external dependency worktrees, `wrk --done` merge-back
(including external cascade), `wrk --list`, `wrk --status`, and project persistence
(`wrk --projects`, `wrk --add`, `wrk --rm`, auto-record, events.jsonl).

# DSN (Domain Specific Notion)

- **wrk CLI** — standalone binary; invocation form `wrk [dir] [flags...]`; first non-flag positional argument is optional `<dir>` — when present, effective cwd is the resolved absolute path of `<dir>` (process cwd unchanged); when absent, effective cwd is process cwd; no-args invocation creates a git worktree from the effective checkout and prints the target path on stdout; `wrk --done` merges the linked worktree branch back into main and removes the worktree + branch (`worktree.MergeBack` with `Remove: true`); `wrk --list` runs `git -C <effective-cwd> worktree list` and prints stdout unchanged; `wrk --status` resolves the effective cwd's git toplevel, scans it with `scan_repo.Scan`, and prints one status block per discovered git directory; missing `<dir>` → `wrk: <path> does not exist`.
- **WRK_HOME** — storage root env var (default `~/.wrk`); tests isolate per run at `{WorkRoot}/.wrk`.
- **WRK_DATE** — optional env var (`YYYY-MM-DD`) overriding the run date for deterministic naming; all tests set `WRK_DATE=2026-06-30`.
- **Work root** — temp directory holding source repos and move targets.
- **Naming** — worktree path `{WRK_HOME}/worktrees/{basename}-{token}-{YYYY-MM-DD}[-N]`; branch `{base-branch}-{YYYY-MM-DD}[-N]`; `N` starts at 1 on collision (path exists or branch ref exists). No unsuffixed names without date.
- **token** — `sanitize(base-branch)` for normal branches (`/` → `-` in path only); for detached HEAD, 7-char short commit hash from `git rev-parse --short=7 HEAD` (not literal `HEAD`).
- **Git source** — cwd must be inside a git checkout (main repo, linked worktree, or nested subdirectory); basename resolves from the checkout root when cwd is a linked worktree or nested subpath.
- **wrk --dep** — spawns a dependency worktree under `{consumerTop}/external/` as a worktree of the DEP repo (`git -C <depMain> worktree add`), so it is registered under `<depMain>/.git/worktrees/` (NOT the consumer's); the dep already holds its own objects, so no remote/fetch into the consumer is needed. **Consumer module discovery**: scans the full consumer tree (`gotool/mod/scan.Scan`) for all Go modules — consumers without a root `go.mod` (module lives in a subdirectory like `go-pkgs/`) are supported. **Dep module discovery**: likewise scans the dep tree for all Go modules. Matches each dep module against every consumer module's `require`/`replace` directives; for every matching consumer module, runs `gotool.Replace` + `gotool.Tidy`. Appends `/external` to `.gitignore` when missing, prints external worktree abs path on stdout. Naming: `{basename}-{token}-{WRK_DATE}[-N]` (same rules as create; basename from dep main repo); the branch lives in the dep repo, so the `-N` collision check runs against `depMain` (path under `consumerTop/external/` + branch in dep repo).
- **wrk --done** — resolves checkout root via `ShowToplevel(cwd)`; requires a linked worktree (not main repo); clean worktree; implicit `--rm`. **Cascade**: `scan_repo.Scan(consumerTop)` discovers every git directory under the checkout; for each row where `RepoType == worktree` and `IsLinked(path)` and `path != checkoutRoot`, run `mergeBackExternalWorktree(path)` in scan path order (path-sorted). This covers `external/*` dep worktrees **and** manually linked worktrees elsewhere (e.g. `deps/foo`). Skip `RepoTypeMain` nested repos (no merge-back/delete). Each cascaded worktree is a dep-repo worktree (registered under `<depMain>/.git/worktrees/`), so `MergeBack` resolves its main repo from the worktree's `.git` gitdir (the dep main) and merges the dep branch back into the dep repo (the branch shares the dep's history, so merge-base resolves); this ensures dep work committed on a nested linked worktree is merged back before removal. Relation to dep main: already-included → remove only; ahead/diverged → prompt (TTY / `-y` on TTY / `--confirm-from-stdin` on own worktree only). **Non-TTY pre-flight guard (option A)**: if stdin is not a terminal and any cascaded linked worktree needs ahead/diverged confirmation, reject `--done` before any cascade mutation — `-y` and `--confirm-from-stdin` do not bypass. No force-removal fallback. The consumer's own `checkoutRoot` is excluded (finished by the final `MergeBack` in `runDone`). **Guard**: scan **every** Go module under the checkout (`gotool/mod/scan.Scan`) — main + all sub-modules — and classify each filesystem/local `replace` (`./`, `../`, or absolute path without version) by resolving its target relative to the module's `go.mod` dir: **intra-repo** = target dir exists AND `git -C <target> rev-parse --show-toplevel` equals the consumer's toplevel (a `../../`/`./sub` nested-module reference back into the same repo); **extra-repo** = everything else (`./external/foo` dep worktree, non-existent target, absolute path to another checkout, sibling outside) (`./external/foo` dep worktree, non-existent target, absolute/sibling outside). The guard names the offending `<top>/<m.Dir>/go.mod` file and each `replace <Old> => <New>` directive in its message. Default (no flag): intra-repo → **WARN to stderr and proceed** (exit 0, merge-back runs); extra-repo → **error, block**. `--no-in-module-replace` (opt-in, valid only with `--done`) → **all** local replaces block (fully-strict). A checkout with no `go.mod` at all yields zero modules → guard is a no-op → `--done` proceeds (and the linked-worktree check inside `MergeBack` still runs for a main-repo cwd, producing `not a linked worktree`). Branch relation to main: already-included → remove only; ahead/diverged → prompt then merge/rebase (or `-y` / `--confirm-from-stdin` on own worktree).
- **wrk --list** — runs `git -C <cwd> worktree list`; prints stdout unchanged; cwd must be inside a git work tree (main repo, linked worktree, or nested subpath). Mutually exclusive with no-args create and `--done`.
- **wrk --status** — standalone reporting mode; cwd must be inside a git work tree. Resolves the effective cwd's checkout root with git toplevel discovery, calls `scan_repo.Scan(context.Background(), scan_repo.Options{Roots: []string{Root}})`, and prints every discovered git directory in scan path order. Each block includes `Dir` relative to the initial toplevel (`.` for the root), current branch, short commit hash plus subject, and `Status` as either `clean` or `dirty (<added> added, <changed> changed, <renamed> renamed, <deleted> deleted)`. **Main repo checkout cwd only** (`worktree.IsMainRepo(checkoutRoot)`): the scan block with `Dir: .` (root checkout) also includes `Remote:` — same brief upstream labels as `--projects` (`identical`, `needs push`, `needs pull`, `diverged`, `(no upstream)`); field order is `Dir`, `Branch`, `Commit`, `Status`, `Remote` (no `Worktrees:`). `Remote:` uses local upstream tracking refs by default; with `--fetch`, fetch upstream for the main repo first then compare. Nested independent `RepoTypeMain` sub-repos and **linked worktree blocks** do **not** show `Remote:`. Running from a **linked worktree cwd** omits `Remote:` on all blocks (append phase also skipped). **Linked worktrees only** (`worktree.IsLinked`) also include one-line `Master:` — brief branch-relation label comparing main repo's current branch vs the worktree's current branch via `git.CompareBranches` (`identical`, `needs merge back(+N commit(s))`, `needs fast forward(+N commit(s))`, `diverged(N commit(s))`). Main checkout blocks (other than `Remote:` above) and nested independent `RepoTypeMain` repos do **not** show `Master:`. When stdout is a TTY or `--color` is set, `Status: clean` is green; dirty status uses granular red/grey segments (same rules as `--projects`); `Master:` and `Remote:` values follow `--projects` color rules when applicable. Without color: plain text. **Append phase (main repo only)**: when `worktree.IsMainRepo(checkoutRoot)`, after all scan blocks (blank line between every block), append one block per **external** linked worktree from `worktree.ListLinked(mainRepo)` in porcelain order, skipping paths already in scan (`scanPaths` dedup — in-tree linked wts like `myrepo/wt-linked` are scan-only). Appended healthy blocks use absolute normalized `Dir` (`storage.NormalizePath`) with full fields (`Branch`, `Commit`, `Status`, `Master:`). Appended **broken** (alive checkout, git fails) blocks are minimal: `Dir` + `Status: error: <git stderr>` only (red `error: …` when `--color`/TTY; run continues). Appended **prunable** (`worktree.IsDead`) blocks are minimal: `Dir` + `Status: prunable` only (plain text). Running from a linked worktree cwd skips the append phase entirely. Mutually exclusive with `--done`, `--list`, `--dep`, `--all-deps`, create target arguments, and other modes.
- **-y / --yes** — universal top-level bool flag; no-op on commands without Y/n prompts (create, `--list`, basename `Select [1-N]`, etc.). Auto-confirms `Proceed? [Y/n]` on own-worktree `--done` / `--merge-back` (stdin check) and `--set-task` rename (stdout check) without reading stdin. On non-TTY, own-worktree ahead/diverged merge-back succeeds with `-y` (no `--confirm-from-stdin` needed). **Cascade guard**: when any cascaded linked worktree needs ahead/diverged confirmation and stdin is non-TTY, `--done` is rejected before mutations — `-y` and `--confirm-from-stdin` do not apply to cascaded worktrees. On TTY, `-y` auto-confirms both consumer and cascaded ahead/diverged worktrees. Recorded in `events.jsonl` `args` when passed.
- **--confirm-from-stdin** — when set with piped `StdinInput`, reads Y/n from stdin for **own-worktree** merge-back confirmation on non-TTY ahead/diverged cases. Superseded by `-y` when `-y` is set (no stdin read). Does **not** confirm cascaded ahead/diverged worktrees on non-TTY (option A pre-flight guard).
- **--no-in-module-replace** — bool flag (no value); valid ONLY with `--done`. Restores the fully-strict local-replace guard: every filesystem/local `replace` (intra-repo or extra-repo) blocks `--done`. Without it (default), intra-repo replaces — whose target dir exists and shares the consumer's `git rev-parse --show-toplevel` (`../../`/`./sub` nested-module reference) — only WARN and `--done` proceeds; extra-repo replaces (`./external/foo` dep worktree, non-existent/absolute/sibling) still block. Bare `wrk --no-in-module-replace`, or with any other mode (`--dep`/`--list`/no-args create/`--all-deps`) → non-zero exit, stderr `wrk: --no-in-module-replace is only valid with --done`.
- **Request.Args** — CLI arguments passed to `wrk` after optional `<dir>` (empty → no-args create; `["--dep", depPath]` for dep tests; `["--done"]` or `["--done", "--confirm-from-stdin"]` for done tests; `["--list"]` for list tests).
- **Request.TargetDir** — when set, prepended as the first positional argument to `wrk` (`wrk <dir> ...`); used by `dir-arg/` tests to run from `WorkRoot` while targeting a repo elsewhere.
- **Request.SpawnDir** — optional second positional `<target-dir>` (`wrk <dir> <target-dir>`); appended after `TargetDir` when set. Overrides the worktree spawn location: missing target with existing parent → spawn exactly at `<target-dir>` (no naming suffix on the path); existing target dir → spawn a default-named sub-dir under it; missing parent → error. Resolved relative to the process (shell) cwd, not `<dir>`. Create-only — errors with `wrk: unexpected arguments` when combined with `--list`/`--done`/`--dep`. `WRK_HOME` is ignored when set.
- **external/** — dependency worktrees live at `{consumerTop}/external/{basename}-{token}-{WRK_DATE}[-N]`; not under `WRK_HOME`. They are linked worktrees of the DEP repo (registered under `<depMain>/.git/worktrees/`), not the consumer — the consumer only hosts the working tree on disk.
- **deps/** — manually linked worktrees may also live under `{consumerTop}/deps/...` (or any path under the checkout); created via `git -C <depMain> worktree add` into the consumer tree. `--done` cascade discovers them via `scan_repo.Scan`, same as `external/*`.
- **runGitIsolated** / **gitOutputIsolated** / **gitWorktreeListIsolated** — thin wrappers over `github.com/xhd2015/gitops/git/git_isolated` (`MustRun`, `MustOutput`, `WorktreeList`).
- **git_isolated** — hook-free git runner; ignores global/system gitconfig; repo-local `core.hookspath` still applies.
- **Session fixtures** — doctest injects `DOCTEST_SESSION_ID` (referenced directly in harness code, not via `os.Getenv`). Seeds live at `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/seeds/` (macOS `cp -c`, Linux `cp -a`); `getWrkBin` builds once per session to `{DOCTEST_FIXTURE_ROOT}/{DOCTEST_SESSION_ID}/bin/wrk` with a file lock across leaf processes.
- **Request.StdinInput** — when non-empty, piped to wrk stdin before wait (mvd merge-back pattern).
- **wrk --all-deps** — automates `--dep` for every dependency discoverable from registered projects: scans the consumer tree (`gotool/mod/scan.Scan`) for all Go modules and builds the union of required-module sets and existing local `Replace` sets; loads candidate dep repos from `{WRK_HOME}/projects.json` via `storage.ListProjects` (lexicographically sorted absolute main-repo paths — same order as `wrk --projects`); for each registered project path that is a git main repo, scans it with `gotool/mod/scan.Scan` and matches module paths against required modules; skips self, modules not required, already-replaced modules (tolerated, not errored — unlike `--done`), already-seen modules, missing project paths, and non-git project paths (all skipped silently); for each match links an external worktree under `{consumerTop}/external/` via the shared `linkExternalDep` core and records a `replace` in every consumer module that requires it (sub-module `Dir` when `m.Dir != "."`); runs `go mod tidy` in each affected consumer module (skipped when zero deps linked). Consumers without a root `go.mod` (module lives in a subdirectory) are supported via module scanning. Stdout: one line per linked dep in **registered project path order** (lexicographic), with matched modules within each project in `mod/scan` Dir order: `wrk <module-path> at ./external/<name>[/<subdir>]` (path relative to `consumerTop`), then a final summary `wrked <N> deps`. Empty or absent `projects.json` → single line `wrked 0 deps`, exit 0, no tidy, no `external/` created, no replaces.
- **Removed from --all-deps** — `--scan-root` flag and `WRK_SCAN_ROOT` env var (and `resolveScanRoot()` / `scan_repo.Scan` in the `--all-deps` code path). Passing `--scan-root` → non-zero exit; stderr mentions unknown/unexpected flag or `scan-root` removed.
- **--all-deps mutual exclusion** — `--all-deps` is mutually exclusive with `--dep`, `--done`, `--list`, and no-args create; `--all-deps --dep <x>` → non-zero exit, stderr mentions "mutually exclusive"; no positional args allowed.
- **wrk --all-deps --dry-run** — runs the full read-only discovery/planning of `--all-deps` but writes nothing. Same cwd/git/go.mod validation, same `required`/`alreadyReplaced`/`consumerModule` sets, same registered-project discovery (`storage.ListProjects` + `modscan.Scan` per project), same self / not-required / already-replaced / seen / missing-path / non-git skips, and the SAME external-path naming + collision logic as the real run — but it does NOT `MkdirAll(external/)`, does NOT `ensureGitignoreExternal`, does NOT `createExternalWorktree` (no `git worktree add`/branch/remote/fetch), does NOT `GoModEditReplace`, and does NOT `GoModTidy`. stdout: one line per planned module in registered project path order `would: wrk <module-path> at ./external/<name>[/<subdir>]`, then a final `would: wrked <N> deps` (empty projects → single `would: wrked 0 deps`). Core guarantee: after a dry run `{consumerTop}/external/` does NOT exist, consumer `go.mod` is unchanged (no new replaces), and `.gitignore` is unchanged (no `/external` line). Errors that occur during planning (non-git cwd, unreadable go.mod) still surface as errors — the process "actually runs".
- **--dry-run** — bool flag (no value); valid ONLY with `--all-deps`. Bare `wrk --dry-run`, or `--dry-run` with any other mode (`--dep`/`--done`/`--list`/no-args create) → non-zero exit, stderr `wrk: --dry-run is only valid with --all-deps`. It does NOT relax `--all-deps`'s mutual exclusion with `--dep`/`--done`/`--list` — `--dry-run --all-deps --dep <x>` still errors as mutually exclusive (the `--all-deps` mutual-exclusion check runs first).
- **wrk --task <desc>** / **wrk -t <desc>** — `-t` and `--task` are equivalent (like `-l,--list`); `hasArg` / `taskFlagSet` detect both forms. Event `args` record whichever form was passed (e.g. `["-t", "desc"]` when `-t` is used). Flag valid only in create mode (no `--done`/`--list`/`--dep`/`--all-deps`). Derives a sanitized slug from `<desc>` (lowercase, non-letter-digit → `-`, collapse, trim, truncate 64 runes). Appends slug after the date in both dir and branch names: `{basename}-{token}-{date}-{slug}[-N]` for dir, `{branchBase}-{date}-{slug}[-N]` for branch. Empty `<desc>` or slug → non-zero exit. No metadata file stored — the slug is embedded in the name.
- **wrk --set-task <desc>** — flag valid alone (mutually exclusive with all other flags). When run from inside a linked worktree (no `<dir>`), renames that worktree. When run as `wrk <dir> --set-task <desc>`, renames the worktree at `<dir>`. Parses the worktree's branch name to extract `branchBase` and `date` (branch must match `{branchBase}-{YYYY-MM-DD}[-slug][-N]`; non-wrk worktrees → error). Computes new dir and branch names with the new slug. If slug is unchanged → no-op. Before `git worktree move`, checks stdout: TTY → warns (old→new path + branch) and prompts `Proceed? [Y/n]`; confirmation executes the move. Non-TTY → non-zero exit `wrk: --set-task requires a terminal (tty)`. When run with `WRK_SET_TASK_CONFIRM=1` env → auto-confirms without TTY (test escape hatch). `<dir>` resolves to the effective working directory; empty `<dir>` (or absent) defaults to cwd.
- **Request.TaskDesc** — when set, task description passed to wrk (with `Request.TaskFlag`, default `--task`).
- **Request.TaskFlag** — task tests: CLI flag form for task description (`-t` or `--task`; default `--task` when `TaskDesc` is set).
- **Request.SetTaskDesc** — when set, tests pass `--set-task <desc>` to wrk; test assertions verify rename side effects.
- **Request.SetTaskEnv** — when set, appended to wrk's environment (e.g., `WRK_SET_TASK_CONFIRM=1` to auto-confirm rename in tests).
- **WRK data storage** — under `{WRK_HOME}`: `projects.json` (recorded main repos) and `events.jsonl` (append-only event log); tests isolate at `{WorkRoot}/.wrk`.
- **Project record** — absolute path to the **main repo** (never a linked worktree path); deduplicated by normalized absolute `path`; fields `path`, `added_at` (ISO-8601 UTC), `source` (`auto` or `manual`); re-adding is idempotent (no duplicate entries; first `source` wins).
- **Auto-record** — on **every** `wrk` invocation, after resolving the effective work directory: if dir missing → no record; if not inside git → no record; otherwise resolve to main repo via `worktree.ResolveMainRepo()` and append to `projects.json` with `source: "auto"` if not already present. Auto-record runs even when the command fails later; failed commands still append an event.
- **WRK_PROJECTS_PERF_LOG** — when set to a file path, `wrk --projects` appends JSONL latency events (`run_start`, `project_start`, `phase`, `worktree_status`, `phase_total`, `project_end`, `run_end`) without changing stdout/stderr; zero overhead when unset.
- **Request.ProjectsPerfLog** — perf-profile tests: path written to `WRK_PROJECTS_PERF_LOG`.
- **wrk --projects** — standalone mode; mutually exclusive with all other modes; prints one **detailed status block** per recorded main repo, sorted lexicographically by absolute path, with blank lines between blocks. **Never aborts** the run due to per-project or per-worktree git failures (exit 0 unless `projects.json` is unreadable); errors surface inline in stdout blocks; stderr stays empty for these cases (unless `-v` is set). **Default (no `--fetch`)**: skip `git fetch`; `Remote:` uses `git.CompareBranches` against local upstream tracking refs. **With `--fetch`**: run scoped upstream fetch (`gitFetchUpstreamQuietNoOptionalLocks`) before `Remote:` comparison per project. **Healthy main repo** blocks include absolute `Dir`, `Branch`, `Commit`, `Status` (same fields as `--status` for the main repo), plus `Remote:` (brief upstream sync summary via `git.CompareBranches`: `identical`, `needs push(+N commit(s))`, `needs pull(N commit(s) behind)`, `diverged(N commit(s))`, `(no upstream)` when the branch has no upstream, or `error: ...` inline when fetch/compare fails), and `Worktrees:` (four spaces after colon, aligned with other fields) with composable summary segments: `N total` and `M dirty` always; `K error` only when K > 0 (alive linked worktree path exists but `git status` fails); `P prune` only when P > 0 (registered in `git worktree list` but checkout directory missing per `worktree.IsDead`). After the `Worktrees:` line, each broken (alive, git-fails) worktree emits `  <absolute-path>  error: <full git stderr message>` (two-space indent); no per-path lines for prunable/dead worktrees. **Broken main repo** blocks omit Branch, Commit, Remote, and Worktrees entirely — only `Dir:` and `Status:       error: <full git stderr message>`. When stdout is a TTY or `--color` is set, highlights attention-worthy **value** portions only: red for the word `dirty`, each dirty count segment with N > 0, `Remote: diverged(...)`, `N dirty` when N > 0, `K error` when K > 0, broken-main `Status: error: ...` value, and per-worktree `error: ...` detail values; grey (`#90`) for dirty count segments with N = 0; orange (`#33`) for `needs push(...)` and `needs pull(...)`; separators `(`, `, `, `)` in dirty status lines stay uncolored; `clean`/`identical`/no-upstream/zero-dirty stay plain (no green on `--projects`). No `<dir>` required; exit 0 when empty (no output). Note: `needs merge back(+N commit(s))` and `needs fast forward(+N commit(s))` apply only to `--status` `Master:` (not `Remote:`).
- **--fetch** — bool flag (no value); valid ONLY with `--projects` or `--status`. Default false (no network fetch). Bare `wrk --fetch`, or `--fetch` with any other mode (`--list`/`--done`/`--dep`/no-args create/etc.) → non-zero exit, stderr `wrk: --fetch is only valid with --projects or --status`. With `--projects` or `--status` from **main repo checkout cwd**: run scoped upstream fetch before `Remote:` comparison. From **linked worktree cwd** with `--status --fetch`: silently ignored (no fetch, no error, no `Remote:` added). Combinable with `--color`. Recorded in `events.jsonl` `args` when passed.
- **-v / --verbose** — global bool flag; valid with **any** wrk mode; does not change mode selection or stdout content. When set, log **major** git subprocesses (mutating/network: `worktree add`/`remove`/`move`, `fetch` when executed, `checkout`, `branch` `-D`/`-m`/`-b`, `merge`, `rebase`, `stash`) to **stderr** as one line per invocation before the command runs: `[YYYY-MM-DD HH:MM:SS] $ git <args...>` (local timezone, format `2006-01-02 15:04:05`; include `-C <dir>` when used). **Not logged**: read-only introspection (`rev-parse`, `log`, `status`, `diff`, `merge-base`, `rev-list --count`, `worktree list`, `show-toplevel`, `config`, etc.) and non-git commands. When `-v` is off: zero stderr logging overhead. Recorded in `events.jsonl` `args` when passed.
- **--color** — bool flag (no value); valid with any mode; forces ANSI coloring on `--projects` and `--status` output even when stdout is a pipe (doctest-safe); no-op on other modes today (e.g. `--list --color` unchanged).
- **Stdout trailing newline** — all wrk modes that print non-empty stdout end with `\n` after the last content line (shell prompt stays on its own line). Empty stdout has no bytes.
- **Stdout assertions** — doctest leaves use `assert.Output` with `version: 2` full-match templates only (no `<contains>` for stdout). Multi-block stdout (e.g. `--status` scan blocks, `--projects` project blocks) is asserted with one v2 template covering the entire stdout; blocks are joined with `\n\n`. Stderr error messages continue to use `<contains>` partial match.
- **wrk --projects streaming** — stdout must flush each lex-ordered project block as soon as that project's gather completes (not after all projects finish). `output-streaming/fast-before-slow-gather` probes pipe timing: first bytes are the fast `aaa` block while the slow `zzz` project (12 worktrees) is still gathering.
- **Run profile labels** — seven leaves are labeled `slow` (>10s cold: 12-worktree perf fixtures, multi-repo `--projects`, linked `--list`, output-streaming probe); `many-worktrees-parallel` is also `flaky` (timing budget). Discovery runs (`doctest test ./tests`) skip labeled leaves; run them with `doctest test --label slow ./tests`.
- **wrk --add `<dir>`** — standalone mode; `--add` consumes the next argument as `<dir>`; validates dir exists + is git; resolves to main repo root; records with `source: "manual"` (idempotent); mutually exclusive with other modes; prints resolved main repo path on stdout (single line) on success.
- **wrk --rm `<dir>`** — standalone mode; `--rm <dir>` (no `--remove` alias); `--rm` consumes the next argument as `<dir>`; mutually exclusive with all other modes; requires non-empty path (`wrk: --rm requires a path argument`). Help text: `--rm <dir>  remove a recorded main repository path`. Resolves target: `filepath.Abs` + `storage.NormalizePath`; if path exists and is inside a git work tree → resolve to main repo via `worktree.ShowToplevel` + `worktree.ResolveMainRepo` (same as `--add`); if path does not exist → use normalized absolute path as-is (stale/moved entries). **Success (entry removed)**: exit 0; stdout = removed main-repo absolute path (single line, trimmed). **Idempotent (not in projects.json)**: exit 0; empty stdout; no error. Does not delete worktrees, git repos, or events.jsonl history. Appends event `command: "rm"`, `args: ["--rm", "<path-arg>"]`, `exit_code: 0`. Auto-record still runs before remove.
- **RemoveProject** — storage API `RemoveProject(wrkHome, path string) (removed bool, err error)` deletes the `projects.json` entry matching normalized absolute `path`; returns whether an entry was removed.
- **events.jsonl** — one JSON object per line appended on every wrk invocation (success or failure): `ts` (ISO-8601 UTC), `command` (mode: `create`, `done`, `list`, `status`, `dep`, `all-deps`, `merge-back`, `set-task`, `repos`, `projects`, `add`, `rm`), `work_dir` (resolved effective cwd), `main_repo` (resolved main repo or empty), `args` (remaining CLI flag args, not positionals), `exit_code`.
- **Request.SecondRepo** — projects tests: second main repo path for multi-project list assertions.
- **Basename fallback** — shared `resolveDirArg` core (`filepath.Abs` → `stat` → optional `projects.json` lookup via `isBasename` / `resolveBasenameFromProjects` / `pickAmbiguousBasename`). When the user-supplied directory argument is a basename (no path separator, not absolute), `stat(filepath.Abs(<dir>))` fails, and `stat(filepath.Join(cwd, <dir>))` also fails: load `projects.json`, collect entries where `filepath.Base(project.path) == <dir>`. **0** → unchanged `wrk: <candidate> does not exist`; **1** → use that project's `path` as the resolved absolute path; **2+** → TTY prints numbered list (candidates sorted lexicographically by absolute path) and prompts `Select [1-N]:`; non-TTY errors listing all candidates. **Skipped** when: `./<dir>` exists in cwd (even non-git — use cwd path, existing git error); or `<dir>` contains a path separator. **Enabled** for: create-mode first positional `<dir>` (`wrk <dir>`, `wrk <dir> <target-dir>`) via `resolveSourceWorkDir` with `allowBasenameFallback=createMode`; and `--dep <dir>` via `runDep` with `allowBasenameFallback=true`. **Not enabled** for other modes (`--list`, `--done`, `--all-deps`, `--status`, `--projects`, `--add`, `--set-task`, `--merge-back`) — positional basename in those modes still skips lookup.
- **WRK_BASENAME_CONFIRM** — when set with piped `StdinInput`, bypasses TTY detection for ambiguous-basename prompt tests (same escape hatch pattern as `WRK_SET_TASK_CONFIRM`).
- **Request.BasenameEnv** — basename-fallback tests: extra env var appended when running wrk (e.g. `WRK_BASENAME_CONFIRM=1`).
- **Request.SelectedSavedRepo** — basename-fallback tty-select: absolute path of the saved project chosen via stdin index.
- **Request.FakeHome** — git-lfs-hook tests: temp home directory holding `$HOME/.local/bin/git-lfs` shim.
- **Request.UseMinimalPath** — when true, wrk runs with `PATH=/usr/bin:/bin` and `HOME={FakeHome}`; git-lfs hook failure is expected (exit 1).

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
│   ├── detached-head/            # cwd on detached HEAD → 7-char hash token
│   └── git-lfs-hooks/            # LFS post-checkout hook requires git-lfs on PATH
│       ├── minimal-path-succeeds/  # stripped PATH; git-lfs in $HOME/.local/bin → create fails
│       └── from-other-cwd/         # wrk <repo> from foreign cwd + stripped PATH → create fails
├── dep/                          # wrk --dep external dependency worktree
│   ├── basic/                    # require + --dep → external wt, replace, tidy, gitignore
│   ├── gitignore-already/        # /external already in .gitignore → no duplicate
│   ├── not-a-dependency/         # dep not in go.mod → error
│   ├── not-git-repo/             # dep path not git → error
│   ├── not-go-module/            # dep git without go.mod → error
│   ├── consumer-no-modules/      # consumer has zero go.mod files → error
│   ├── dep-sub-module/           # dep root has no go.mod; module in subdir consumer requires → external wt + replace => <external>/sub
│   ├── dep-multi-sub-module/     # dep has multiple sub-modules; match correct one → replace => <external>/b
│   ├── consumer-sub-module/      # consumer has no root go.mod; module in subdir → scan + replace in sub-module
│   ├── consumer-multi-module/    # consumer has two sub-modules both requiring dep → replace in both
│   ├── consumer-multi-module-selective/ # consumer has two sub-modules, only one requires dep → replace only in that one
│   ├── both-sub-modules/         # both consumer + dep go.mod in subdirs → replace => <external>/sub
│   ├── cwd-in-sub-module/        # cwd inside sub-module dir, not repo root → success
│   ├── external-wt-owned-by-dep-repo/  # external wt .git gitdir points into dep main, not consumer
│   ├── external-wt-from-linked-consumer/ # --dep from inside a linked consumer wt → owned by dep main, not consumer main
│   └── basename-fallback/        # wrk --dep <basename> → saved projects.json lookup (same core as create)
│       ├── single-match/
│       │   └── basic/            # one saved dep match → external wt + replace
│       ├── cwd-exists/
│       │   └── no-fallback/        # ./mydep in consumer cwd → local path, no lookup
│       ├── path-with-separator/
│       │   └── no-fallback/        # --dep sub/mydep missing → no lookup
│       ├── no-match/
│       │   └── error/            # zero matches → does not exist
│       └── ambiguous/
│           ├── tty-select/       # WRK_BASENAME_CONFIRM + stdin selects dep
│           └── non-tty/          # error listing candidates
├── all-deps/                     # wrk --all-deps: registered projects.json → link each matched dep
│   ├── registered/               # dep discovery from projects.json (not blind scan)
│   │   ├── basic/              # dep1+dep2 registered → both linked, project-path order, 2 deps
│   │   ├── partial/            # only dep1 registered → dep1 linked, 1 deps
│   │   ├── nested-submodule/   # nested sub-module in registered repo → one wt, replace at sub-dir
│   │   ├── multi-module-same-repo/ # two sub-modules same repo → one wt, two replaces, 2 deps
│   │   ├── self-skip/          # consumer also registered → dep linked, self skipped, 1 deps
│   │   ├── not-a-dep/          # registered project not required → skipped, 0 deps
│   │   ├── empty-projects/     # no projects.json → 0 deps, no external/, no replaces
│   │   ├── already-replaced/   # dep1 pre-replaced → skipped; dep2 linked, 1 deps
│   │   ├── missing-project-path/ # projects.json path missing → skip silently, 0 deps
│   │   ├── non-git-project/    # registered non-git dir → skip silently, 0 deps
│   │   ├── consumer-sub-module/  # consumer module in subdir → scan + link works
│   │   ├── removed-scan-root/  # --scan-root flag → non-zero error
│   │   └── dry-run/            # --dry-run with registered projects
│   │       ├── basic/          # registered deps → would: lines, no side effects
│   │       └── empty/        # empty projects → would: wrked 0 deps
│   ├── mutually-exclusive/     # wrk --all-deps --dep <x> → non-zero, mutually exclusive
│   ├── not-git-cwd/            # cwd not a git repo → non-zero, is not a git repository
│   └── dry-run/
│       └── without-all-deps/   # wrk --dry-run (no --all-deps) → non-zero
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
│   ├── cascade-force-removal/    # bug: non-TTY ahead cascade must error, not force-remove
│   │   ├── ahead-non-tty-errors/
│   │   └── ahead-non-tty-with-y-still-errors/
│   ├── cascade-non-tty-rejects-with-confirm-from-stdin/ # option A: --confirm-from-stdin cannot confirm cascade on non-TTY
│   ├── cascade-dep-merge-back/   # regression: ahead dep + --confirm-from-stdin on non-TTY → pre-flight error (no cascade merge)
│   ├── cascade-non-external-linked/ # manual deps/foo linked wt (not under external/) → cascade removes it, consumer merge-back exit 0
│   ├── cascade-external-and-deps/ # external/* + deps/foo both linked → cascade removes both, consumer merge-back exit 0
│   ├── local-replace-blocks/     # extra-repo fs replace (non-existent ./external/foo) → guard blocks + names go.mod + directive
│   ├── intra-replace-warns/      # intra-repo fs replace (./submod, same toplevel) → WARN + proceed (default, exit 0)
│   ├── intra-replace-cross-worktree/ # abs replace to sibling checkout; wrk --done <wt> from outside → extra-repo block
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
├── fetch-and-verbose/            # --fetch opt-in refresh + -v verbose git logging + --status Remote:
│   ├── fetch/
│   │   ├── invalid-mode/       # --fetch without --projects/--status → error
│   │   │   ├── bare/
│   │   │   ├── with-list/
│   │   │   └── with-done/
│   │   ├── projects/           # default no-fetch vs --fetch on --projects
│   │   │   ├── default-no-fetch/
│   │   │   │   └── stale-tracking-ref/
│   │   │   └── with-fetch/
│   │   │       ├── behind-upstream/
│   │   │       └── fetch-failure/
│   │   └── status/             # --fetch on --status (main vs linked cwd)
│   │       ├── stale-ref-no-fetch/
│   │       ├── with-fetch-behind/
│   │       ├── from-linked-ignore-fetch/
│   │       ├── from-linked-fetch-not-run/
│   │       └── with-fetch-logs/
│   ├── verbose/                # -v major-git-command stderr logging
│   │   ├── list/
│   │   │   └── no-log/         # worktree list is minor → empty stderr
│   │   ├── create/
│   │   │   ├── basic/          # worktree add logged
│   │   │   └── no-minor/       # no rev-parse/status lines
│   │   ├── projects/
│   │   │   ├── no-fetch/       # minor reads only → empty stderr
│   │   │   └── with-fetch/     # fetch logged
│   │   ├── done/
│   │   │   └── merge-back/     # merge/worktree remove logged
│   │   └── off/
│   │       └── no-stderr/      # no -v → empty stderr
│   └── status/
│       └── remote/             # --status Remote: on main checkout root block
│           ├── main-clean/
│           │   └── identical/
│           ├── main-no-upstream/
│           └── from-linked-no-remote/
├── status/                       # wrk --status status-block display
│   ├── valid-git-cwd/            # cwd resolves to a git checkout
│   │   ├── root-clean/           # root checkout shown as "." and clean
│   │   ├── subdir-clean/         # nested cwd still reports root-relative "."
│   │   ├── multiple-git-dirs/    # root + nested independent repo blocks
│   │   └── dirty-counts/         # added/changed/renamed/deleted counts
│   ├── invalid-git-cwd/
│   │   └── non-git/              # cwd is not a git repo (error)
│   ├── master-field/             # brief Master: on linked worktrees only (plain pipe)
│   │   ├── linked-ahead/         # Master: needs fast forward(+N commits)
│   │   ├── linked-identical/     # Master: identical
│   │   ├── linked-merge-back/    # Master: needs merge back(+N commits)
│   │   ├── linked-diverged/      # Master: diverged(N commits)
│   │   ├── main-no-compare/      # main checkout omits field
│   │   └── nested-main-no-compare/ # nested independent repo omits field
│   ├── color-output/             # wrk --status alignment + conditional ANSI (--color)
│   │   ├── force-color-clean/    # --color → green Status: clean
│   │   ├── force-color-dirty/    # --color → granular red/grey dirty status
│   │   ├── force-color-master-identical/   # green Master: identical
│   │   ├── force-color-master-fast-forward/ # orange needs fast forward
│   │   ├── force-color-master-merge-back/   # orange needs merge back
│   │   ├── force-color-master-diverged/   # red diverged
│   │   └── no-color-pipe/        # pipe without --color → no ANSI, brief Master:
│   └── invalid-mode/
│       └── with-list/            # --status with --list is mutually exclusive
│   └── main-repo-worktrees/      # nested DOCTEST: append external linked wts from main repo
│       ├── no-linked-external/   # clean main, no external wt → scan only
│       ├── external-clean/       # wrk external → scan + appended full block
│       ├── external-dirty/       # external wt dirty counts in append
│       ├── in-tree-only/         # in-tree git worktree add only → no append
│       ├── mixed-external-in-tree/ # scan in-tree + append external only
│       ├── external-broken/      # alive path, broken git → minimal error block
│       ├── external-prunable/    # removed checkout → minimal prunable block
│       ├── from-linked-cwd/      # --status inside external wt → no append
│       ├── ordering-two-external/ # two external wts → ListLinked order
│       └── color-broken/         # --color red error on appended broken block
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
        │   ├── t-alias/               # wrk -t "fix login bug" → slug in name; event args use -t
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
            ├── not-wrk-worktree/      # custom branch → cannot parse → error
            ├── rename-succeeds/       # TTY-confirmed rename via WRK_SET_TASK_CONFIRM=1
            ├── slug-unchanged/        # same slug → no-op "task unchanged"
            ├── propagate/             # --set-task updates gitdir for nested repos
            │   ├── single-external-dep/ # external dep's gitdir updated to new path
            │   ├── non-external-linked-dep/ # manual deps/foo linked wt gitdir updated to new path
            │   └── abs-replace-rewritten/ # go.mod abs replace rewritten after consumer rename
            └── with-dir/              # wrk <dir> --set-task (target via argument)
                ├── rename-succeeds/   # rename worktree at given <dir>
                ├── empty-desc/        # empty description → error
                ├── mutually-exclusive/# with --list → mutual exclusion error
                └── missing-dir/       # non-existent dir → "does not exist"
├── yes-flag/                     # universal -y / --yes flag
│   ├── done/
│   │   ├── ahead-non-tty/        # wrk --done -y on own ahead wt (non-TTY)
│   │   └── ahead-no-prompt/      # TTY + -y: no Proceed? shown (label: tty)
│   ├── merge-back/
│   │   └── ahead-non-tty/        # wrk --merge-back -y merges, keeps wt
│   ├── set-task/
│   │   └── rename-non-tty/       # wrk --set-task -y renames without TTY error
│   ├── no-op/
│   │   └── create-with-yes/      # wrk -y create same as bare wrk
│   └── cascade/
│       ├── non-tty-rejects/      # ahead external + wrk --done -y → error
│       └── tty-auto-yes/         # TTY + -y auto-confirms cascade merge (label: tty)
└── projects/                     # project persistence + event logging
    ├── auto-record/              # auto-record main repo on every invocation
    │   ├── no-dir/               # effective work dir = process cwd
    │   │   ├── main-cwd/         # cwd is main repo root
    │   │   └── subdir-cwd/       # cwd is nested subpath inside repo
    │   ├── dir-arg/              # effective work dir = <dir> positional
    │   │   ├── main-repo/        # wrk <mainRepo>
    │   │   ├── linked-worktree/  # wrk <linkedWt> → main repo
    │   │   └── nested-subpath/   # wrk <nestedSubpath> → main repo
    │   ├── non-git/              # non-git cwd → no record
    │   ├── missing-dir/          # wrk <nonexistent> → no record
    │   └── fail-after-record/    # dirty --done fails but project recorded
    ├── remote-brief/             # wrk --projects shared Remote: brief labels (plain pipe)
    │   ├── ahead-of-upstream/    # Remote: needs push(+N commit)
    │   ├── behind-upstream/      # Remote: needs pull(N commit(s) behind)
    │   ├── diverged/             # Remote: diverged(N commits)
    │   └── up-to-date/           # Remote: identical
    ├── detailed-status/          # wrk --projects detailed status blocks (plain pipe output)
    │   ├── single-clean-no-wts/  # one project, clean, no linked wts
    │   ├── stale-gitdir-linked/  # stale .git gitdir -> inline error detail, exit 0
    │   ├── broken-main-repo/     # recorded path no longer git -> minimal Dir+Status error block
    │   ├── prunable-worktrees/   # deleted checkout -> 0 total, 1 prune (summary only)
    │   ├── with-linked-mixed/    # Worktrees:    3 total, 1 dirty
    │   ├── ahead-of-upstream/    # Remote: needs push(+N commit)
    │   ├── no-upstream/          # Remote: (no upstream)
    │   ├── multiple-projects/    # two blocks, lex order, blank separator
    │   └── empty/                # exit 0, empty stdout
    ├── color-output/             # wrk --projects alignment + conditional ANSI (--color)
    │   ├── no-color-pipe/        # pipe without --color → no ANSI, aligned Worktrees
    │   ├── force-color-dirty-status/   # granular red/grey dirty Status segments
    │   ├── force-color-dirty-partial/  # 2 changed, zero other counts → grey + red mix
    │   ├── force-color-needs-push/     # orange needs push(...)
    │   ├── force-color-needs-pull/     # orange needs pull(...)
    │   ├── force-color-diverged/       # red diverged(...)
    │   ├── force-color-worktrees-dirty/ # red N dirty portion only
    │   ├── force-color-stale-gitdir-linked/ # red on error summary + detail lines
    │   ├── clean-no-color/       # all clean + --color → no highlights
    │   └── color-with-list/      # --list --color → list unchanged, no ANSI
    ├── list/
    │   └── projects/
    │       ├── empty/            # wrk --projects empty → exit 0, no output
    │       └── after-records/    # sorted detailed blocks after auto-record
    ├── add/
    │   ├── manual/
    │   │   ├── main-repo/        # wrk --add <mainRepo>
    │   │   └── linked-worktree/  # wrk --add <linkedWt> → main repo
    │   └── idempotent/           # auto + manual → single entry
    ├── remove/
    │   ├── manual/
    │   │   ├── main-repo/        # --add then --rm <mainRepo> → gone, stdout path
    │   │   └── linked-worktree/  # --rm <linkedWt> → main repo removed
    │   ├── idempotent/
    │   │   ├── not-recorded/     # never recorded → exit 0, empty stdout
    │   │   └── already-removed/  # remove twice → second empty stdout
    │   ├── missing-path-arg/     # wrk --rm (no path) → error
    │   ├── stale-path/           # record, delete .git, --rm old path → removed
    │   └── invalid-mode/
    │       └── remove-with-list/ # wrk --rm X --list → mutually exclusive
    ├── events/
    │   ├── append-on-success/    # create → event exit_code 0
    │   └── append-on-failure/    # failed command → event exit_code != 0
    ├── invalid-mode/
    │   ├── projects-with-list/   # wrk --projects --list → mutual exclusion
    │   └── add-missing-path/     # wrk --add without path → error
    ├── output-streaming/         # wrk --projects incremental stdout (per-project as ready)
    │   └── fast-before-slow-gather/ # fast aaa block streams before slow zzz gather ends
    ├── perf-profile/             # WRK_PROJECTS_PERF_LOG instrumentation + parallel budgets
    │   ├── emits-events/
    │   │   └── many-worktrees/   # JSONL lifecycle + 12 worktree_status events
    │   ├── budget/
    │   │   └── many-worktrees-parallel/ # worktree_status_all <100ms, run_end <200ms
    │   └── structure/
    │       └── dedup-list-linked/  # single ListLinked per project (not skip+summary)
    └── basename-fallback/        # create-mode basename → saved projects.json lookup
        ├── single-match/
        │   └── create/           # one match → worktree from saved path
        ├── cwd-exists/
        │   └── no-fallback/      # ./basename in cwd (non-git) → no lookup
        ├── no-match/
        │   └── error/            # zero matches → does not exist
        ├── path-with-separator/
        │   └── no-fallback/      # sub/foo → no lookup
        ├── ambiguous/
        │   ├── tty-select/       # WRK_BASENAME_CONFIRM + stdin index
        │   └── non-tty/          # error listing candidates
        └── other-mode/
            └── no-fallback/      # wrk basename --list → no lookup
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
| 8a | create-worktree/git-lfs-hooks/minimal-path-succeeds | LFS hook + stripped PATH; git-lfs in $HOME/.local/bin → create fails (expected) |
| 8b | create-worktree/git-lfs-hooks/from-other-cwd | wrk \<repo\> from foreign cwd + stripped PATH → create fails (expected) |
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
| 25 | dep/consumer-no-modules | Consumer has zero go.mod files → error |
| 26 | dep/dep-sub-module | Dep root has no go.mod; module in subdir that consumer requires → external wt + replace => `<external>/sub` |
| 27 | dep/dep-multi-sub-module | Dep has multiple sub-modules; consumer requires one → match correct one, replace => `<external>/b` |
| 28 | dep/consumer-sub-module | Consumer has no root go.mod; module in subdir → scan + replace in matching sub-module |
| 29 | dep/consumer-multi-module | Consumer has two sub-modules both requiring dep → replace in both |
| 30 | dep/consumer-multi-module-selective | Consumer has two sub-modules, only one requires dep → replace only in that one |
| 31 | dep/both-sub-modules | Both consumer + dep go.mod in subdirs → replace => `<external>/sub` |
| 32 | dep/cwd-in-sub-module | Cwd is inside sub-module dir, not repo root → success |
| 33 | dep/external-wt-owned-by-dep-repo | External dep worktree's `.git` gitdir resolves to the DEP main repo, not the consumer; dep owns the worktree |
| 34 | dep/external-wt-from-linked-consumer | `--dep` from inside a linked consumer worktree → external wt owned by dep main repo, not consumer main |
| 35 | all-deps/registered/basic | dep1+dep2 registered in projects.json → both linked in project-path order, `wrked 2 deps` |
| 36 | all-deps/registered/partial | only dep1 registered → dep1 linked, dep2 not replaced, `wrked 1 deps` |
| 37 | all-deps/registered/already-replaced | dep1 pre-replaced `=> ./external/preexisting` → skipped; dep2 linked, `wrked 1 deps` |
| 38 | all-deps/registered/empty-projects | no projects.json → `wrked 0 deps`, no replaces, no `external/` |
| 39 | all-deps/registered/self-skip | consumer also in projects.json → dep1 linked, self skipped, `wrked 1 deps` |
| 40 | all-deps/mutually-exclusive | `wrk --all-deps --dep <x>` → non-zero, mutually exclusive |
| 41 | all-deps/not-git-cwd | cwd not a git repo → non-zero, is not a git repository |
| 42 | all-deps/registered/nested-submodule | nested sub-module in registered repo → linked, replace at sub-dir, `wrked 1 deps` |
| 43 | all-deps/registered/multi-module-same-repo | two sub-modules in one registered repo → ONE worktree, two replaces, `wrked 2 deps` |
| 44 | all-deps/registered/consumer-sub-module | consumer module in subdir → scan + link all registered deps |
| 45 | all-deps/registered/dry-run/basic | registered dep1+dep2 → `would:` lines, `would: wrked 2 deps`, NO side effects |
| 46 | all-deps/registered/dry-run/empty | empty projects → `would: wrked 0 deps`, NO side effects |
| 47 | all-deps/dry-run/without-all-deps | `wrk --dry-run` (no `--all-deps`) → non-zero, stderr `--dry-run is only valid with --all-deps` |
| 47a | all-deps/registered/not-a-dep | registered project not required by consumer → skipped, `wrked 0 deps` |
| 47b | all-deps/registered/missing-project-path | projects.json missing path → skip silently, `wrked 0 deps` |
| 47c | all-deps/registered/non-git-project | registered non-git dir → skip silently, `wrked 0 deps` |
| 47d | all-deps/registered/removed-scan-root | `wrk --all-deps --scan-root X` → non-zero, stderr mentions scan-root |
| 48 | done/external-cascade | `--done` cascades to `external/*` dep wt first; parent errors on local replace (names go.mod + directive) |
| 49 | done/local-replace-blocks | extra-repo `replace => ./external/foo` (non-existent) blocks `--done` at guard (names go.mod + directive) |
| 50 | dir-arg/create/basic | `wrk <repoDir>` from WorkRoot creates worktree |
| 51 | dir-arg/list/from-dir | `wrk <repoDir> --list` matches `git worktree list` |
| 52 | dir-arg/missing-dir | `wrk <nonexistent>` → does not exist |
| 53 | target-dir/target-missing/parent-exists | `wrk <dir> <target>` spawns exactly at `<target>` (parent exists) |
| 54 | target-dir/target-missing/parent-missing | `wrk <dir> <target>` parent missing → error |
| 55 | target-dir/target-exists/basic-subdir | `wrk <dir> <target>` existing dir → default-named sub-dir |
| 56 | target-dir/target-exists/collision-suffix | existing dir + colliding sub-dir → `-N` suffix |
| 57 | target-dir/target-exists/target-is-file | `<target>` is a file → error |
| 58 | target-dir/relative-path | relative `<target>` resolved against shell cwd |
| 59 | target-dir/with-other-mode/with-list | `wrk <dir> <target> --list` → unexpected arguments |
| 60 | target-dir/with-other-mode/with-dep | `wrk <dir> <target> --dep <dep>` → unexpected arguments |
| 61 | done/no-go-mod | linked wt whose checkout has no go.mod; `--done` merge-back succeeds (guard no-op) |
| 62 | done/not-linked-no-go-mod | main repo without go.mod; `--done` → `not a linked worktree` (not `no go.mod found`) |
| 63 | done/sub-module-replace-blocks | main go.mod clean but `sub/go.mod` has local replace → guard blocks `--done` (names go.mod + directive) |
| 64 | done/cascade-merge-base | cascade must remove dep wt, not crash `failed to find merge base` (dep branch vs consumer main share no history) |
| 65 | done/cascade-force-removal/ahead-non-tty-errors | ahead external dep + non-TTY `--done` → error; dep wt + commits preserved (no force-remove) |
| 65a | done/cascade-force-removal/ahead-non-tty-with-y-still-errors | same + `-y` → still errors (cascade guard) |
| 65b | done/cascade-non-tty-rejects-with-confirm-from-stdin | ahead external + `--confirm-from-stdin` on non-TTY → error before mutations |
| 65c | done/cascade-dep-merge-back | regression: ahead dep + `--confirm-from-stdin` on non-TTY → error (option A; no cascade merge) |
| 120 | yes-flag/done/ahead-non-tty | `wrk --done -y` merges own ahead wt on non-TTY |
| 121 | yes-flag/done/ahead-no-prompt | TTY + `wrk --done -y` shows no `Proceed?` (label: tty) |
| 122 | yes-flag/merge-back/ahead-non-tty | `wrk --merge-back -y` merges, keeps worktree |
| 123 | yes-flag/set-task/rename-non-tty | `wrk --set-task -y` renames on non-TTY |
| 124 | yes-flag/no-op/create-with-yes | `wrk -y` create same as bare `wrk` |
| 125 | yes-flag/cascade/non-tty-rejects | ahead external + `wrk --done -y` on non-TTY → error |
| 126 | yes-flag/cascade/tty-auto-yes | TTY + `wrk --done -y` merges cascade + consumer (label: tty) |
| 65a | done/cascade-non-external-linked | manual `deps/foo` linked wt (not under `external/`) → cascade removes it, consumer `--done` exit 0 |
| 65b | done/cascade-external-and-deps | `external/*` + `deps/foo` both linked → cascade removes both, consumer `--done` exit 0 |
| 66 | done/intra-replace-warns | intra-repo `replace example.com/foo => ./submod` (existing, same toplevel) → WARN, exit 0, merge-back proceeds |
| 66b | done/intra-replace-cross-worktree | abs replace to main-checkout `submod` from linked wt; `wrk --done <wt>` from outside → extra-repo block |
| 67 | done/intra-replace-strict-blocks | intra-repo replace + `--no-in-module-replace` → block, names go.mod + directive |
| 68 | done/no-in-module-replace-without-done | `wrk --list --no-in-module-replace` → non-zero, `--no-in-module-replace is only valid with --done` |
| 69 | task/spawn/basic | `wrk --task "fix login bug"` → dir/branch include `-fix-login-bug` |
| 69a | task/spawn/t-alias | `wrk -t "fix login bug"` → same slug behavior; event `args: ["-t", "fix login bug"]` |
| 70 | task/spawn/special-chars | Task with capitals, symbols, unicode → sanitized slug |
| 71 | task/spawn/long-task | >64 runes → truncated to 64 |
| 72 | task/spawn/empty-task | `--task ""` → error |
| 73 | task/spawn/empty-slug | `--task "!!!"` → error (slug empty after sanitization) |
| 74 | task/spawn/with-done | `--task` + `--done` → mutually exclusive |
| 75 | task/spawn/sequence | Two `wrk --task "same"` calls → `-N` suffix on second |
| 76 | task/spawn/branch-collision | Pre-existing branch with task-slug name → suffix increment |
| 77 | task/spawn/target-dir | `wrk <dir> <target> --task "desc"` → branch has slug, dir is user-specified |
| 78 | task/set-task/non-tty | `--set-task` in non-TTY → error "requires terminal" |
| 79 | task/set-task/empty-desc | `--set-task ""` → error |
| 80 | task/set-task/empty-slug | `--set-task "!!!"` → error |
| 81 | task/set-task/not-linked | `--set-task` from main repo → error |
| 82 | task/set-task/not-wrk-worktree | `--set-task` on custom-branch worktree → cannot parse → error |
| 83 | task/set-task/rename-succeeds | `--set-task "new task"` with WRK_SET_TASK_CONFIRM=1 → worktree renamed, branch renamed |
| 84 | task/set-task/slug-unchanged | `--set-task` with same slug → no-op, prints "task unchanged" |
| 85 | task/set-task/propagate/single-external-dep | `--set-task` with external dep → consumer renamed, dep gitdir updated to new path |
| 85a | task/set-task/propagate/non-external-linked-dep | `--set-task` with manual `deps/foo` linked wt → consumer renamed, dep gitdir updated to new path |
| 85b | task/set-task/propagate/abs-replace-rewritten | `--set-task` with `wrk --dep` abs replace → go.mod replace rewritten to new consumer path |
| 86 | task/set-task/with-dir/rename-succeeds | `wrk <dir> --set-task "new task"` renames worktree at `<dir>` |
| 87 | task/set-task/with-dir/empty-desc | `wrk <dir> --set-task ""` → error |
| 88 | task/set-task/with-dir/mutually-exclusive | `wrk <dir> --set-task "task" --list` → mutual exclusion error |
| 89 | task/set-task/with-dir/missing-dir | `wrk <nonexistent> --set-task "task"` → does not exist |
| 83 | status/valid-git-cwd/root-clean | `wrk --status` from repo root shows `Dir: .` and clean status |
| 84 | status/valid-git-cwd/subdir-clean | `wrk --status` from nested subdir still shows `Dir: .` |
| 85 | status/valid-git-cwd/multiple-git-dirs | root + nested independent git repo produce two status blocks |
| 86 | status/valid-git-cwd/dirty-counts | status counts one added, changed, renamed, and deleted entry |
| 87 | status/invalid-git-cwd/non-git | `wrk --status` outside git fails with `is not a git repository` |
| 88 | status/invalid-mode/with-list | `wrk --status --list` fails as mutually exclusive |
| 88a | status/master-field/linked-ahead | linked wt `Master: needs fast forward(+1 commit)` |
| 88b | status/master-field/linked-identical | linked wt `Master: identical` |
| 88c | status/master-field/linked-merge-back | linked wt `Master: needs merge back(+1 commit)` |
| 88d | status/master-field/linked-diverged | linked wt `Master: diverged(2 commits)` |
| 88e | status/master-field/main-no-compare | main checkout block has no Master: line |
| 88f | status/master-field/nested-main-no-compare | nested independent repo has no Master: line |
| 88g | status/color-output/force-color-clean | `--color` → green `Status: clean` |
| 88h | status/color-output/force-color-dirty | `--color` → granular red/grey dirty status |
| 88i | status/color-output/force-color-master-identical | `--color` → green `Master: identical` |
| 88j | status/color-output/force-color-master-fast-forward | `--color` → orange needs fast forward |
| 88k | status/color-output/force-color-master-merge-back | `--color` → orange needs merge back |
| 88l | status/color-output/force-color-master-diverged | `--color` → red diverged |
| 88m | status/color-output/no-color-pipe | pipe `--status` → no ANSI, brief Master: |
| 88n | status/main-repo-worktrees/no-linked-external | main repo only; scan block unchanged, no append |
| 88o | status/main-repo-worktrees/external-clean | scan `.` + appended full external block (abs Dir, Master) |
| 88p | status/main-repo-worktrees/external-dirty | appended external `Status: dirty (...)` |
| 88q | status/main-repo-worktrees/in-tree-only | in-tree linked wt scan-only; no append (dedup) |
| 88r | status/main-repo-worktrees/mixed-external-in-tree | scan in-tree + append external only |
| 88s | status/main-repo-worktrees/external-broken | appended minimal `Status: error: …`, exit 0 |
| 88t | status/main-repo-worktrees/external-prunable | appended minimal `Status: prunable` |
| 88u | status/main-repo-worktrees/from-linked-cwd | `--status` from external wt; no append section |
| 88v | status/main-repo-worktrees/ordering-two-external | two external wts; append order = ListLinked |
| 88w | status/main-repo-worktrees/color-broken | `--status --color` red `error:` on appended block |
| 90 | projects/auto-record/no-dir/main-cwd | `wrk --list` from main repo cwd records main repo |
| 91 | projects/auto-record/no-dir/subdir-cwd | `wrk --list` from nested subdir records main repo |
| 92 | projects/auto-record/dir-arg/main-repo | `wrk <mainRepo> --list` records main repo |
| 93 | projects/auto-record/dir-arg/linked-worktree | `wrk <linkedWt> --list` records main repo, not worktree |
| 94 | projects/auto-record/dir-arg/nested-subpath | `wrk <nestedSubpath> --list` records main repo |
| 95 | projects/auto-record/non-git | non-git cwd → no project record |
| 96 | projects/auto-record/missing-dir | `wrk <nonexistent> --list` → no project record |
| 97 | projects/auto-record/fail-after-record | dirty `--done` fails but project auto-recorded + event logged |
| 98 | projects/list/projects/empty | `wrk --projects` empty → exit 0, no output |
| 99 | projects/list/projects/after-records | `wrk --projects` prints sorted detailed blocks after auto-record |
| 99a | projects/detailed-status/single-clean-no-wts | one project block with remote compare + `0 total, 0 dirty` |
| 99b | projects/detailed-status/with-linked-mixed | `Worktrees: 3 total, 1 dirty` |
| 99b2 | projects/detailed-status/stale-gitdir-linked | stale `.git` gitdir → `2 total, 0 dirty, 1 error` + detail line, exit 0 |
| 99b2a | projects/detailed-status/broken-main-repo | broken main repo → minimal `Dir` + `Status: error: ...` only, exit 0 |
| 99b2b | projects/detailed-status/prunable-worktrees | deleted checkout → `0 total, 0 dirty, 1 prune`, no per-path lines, exit 0 |
| 99c | projects/detailed-status/ahead-of-upstream | `Remote:` shows `needs push(+N commit)` |
| 99d | projects/detailed-status/no-upstream | `Remote: (no upstream)` |
| 99e | projects/detailed-status/multiple-projects | two lex-ordered blocks with blank separator |
| 99e2 | projects/output-streaming/fast-before-slow-gather | fast project stdout before slow project gather completes |
| 99f | projects/detailed-status/empty | empty projects → exit 0, no stdout |
| 99g | projects/color-output/no-color-pipe | pipe `--projects` → no ANSI, aligned `Worktrees:    ` |
| 99h | projects/color-output/force-color-dirty-status | `--color` → granular red/grey dirty status segments |
| 99h2 | projects/color-output/force-color-dirty-partial | `--color` → grey zero segments, red `2 changed` |
| 99i | projects/color-output/force-color-needs-push | `--color` → orange around `needs push(...)` |
| 99j | projects/color-output/force-color-needs-pull | `--color` → orange around `needs pull(...)` |
| 99k | projects/color-output/force-color-diverged | `--color` → red around `diverged(...)` |
| 99o | projects/remote-brief/ahead-of-upstream | plain `Remote: needs push(+1 commit)` |
| 99p | projects/remote-brief/behind-upstream | plain `Remote: needs pull(1 commit behind)` |
| 99q | projects/remote-brief/diverged | plain `Remote: diverged(2 commits)` |
| 99r | projects/remote-brief/up-to-date | plain `Remote: identical` |
| 99l | projects/color-output/force-color-worktrees-dirty | `--color` → red on `N dirty` only |
| 99l2 | projects/color-output/force-color-stale-gitdir-linked | `--color` → red on `1 error` summary + per-path `error: ...` |
| 99m | projects/color-output/clean-no-color | all clean + `--color` → no red/orange on values |
| 99n | projects/color-output/color-with-list | `--list --color` → git worktree list unchanged |
| 100 | projects/add/manual/main-repo | `wrk --add <mainRepo>` records + stdout path |
| 101 | projects/add/manual/linked-worktree | `wrk --add <linkedWt>` resolves to main repo |
| 102 | projects/add/idempotent | duplicate auto + manual → single entry (source stays auto) |
| 103 | projects/events/append-on-success | create appends event with `exit_code` 0 |
| 104 | projects/events/append-on-failure | failed command appends event with `exit_code` != 0 |
| 105a | projects/perf-profile/emits-events/many-worktrees | perf log JSONL with run/project/phase/worktree events for 12 wts |
| 105b | projects/perf-profile/budget/many-worktrees-parallel | parallel gather: worktree_status_all <100ms, run_end <200ms |
| 105c | projects/perf-profile/structure/dedup-list-linked | one list_linked phase per project (dedup ListLinked) |
| 105 | projects/invalid-mode/projects-with-list | `wrk --projects --list` → mutually exclusive error |
| 106 | projects/invalid-mode/add-missing-path | `wrk --add` without path → error |
| 106a | projects/remove/manual/main-repo | `--add` then `--rm <mainRepo>` → gone from json, stdout path |
| 106b | projects/remove/manual/linked-worktree | `--rm <linkedWt>` resolves to main repo, removes entry |
| 106c | projects/remove/idempotent/not-recorded | `--rm` never-recorded path → exit 0, empty stdout |
| 106d | projects/remove/idempotent/already-removed | remove twice → second call exit 0, empty stdout |
| 106e | projects/remove/missing-path-arg | `wrk --rm` without path → error |
| 106f | projects/remove/stale-path | record repo, delete `.git`, `--rm <old-path>` → removed |
| 106g | projects/remove/invalid-mode/remove-with-list | `wrk --rm X --list` → mutually exclusive error |
| 107 | projects/basename-fallback/single-match/create | Saved project; cwd elsewhere; `wrk myrepo` creates wt from saved path |
| 108 | projects/basename-fallback/cwd-exists/no-fallback | `./myrepo` exists in cwd (not git); `wrk myrepo` → git error, no fallback |
| 109 | projects/basename-fallback/no-match/error | No cwd entry, no saved project → `does not exist` |
| 110 | projects/basename-fallback/path-with-separator/no-fallback | `wrk sub/foo` missing → no fallback, normal error |
| 111 | projects/basename-fallback/ambiguous/tty-select | Two saved projects same basename; TTY + stdin selects one |
| 112 | projects/basename-fallback/ambiguous/non-tty | Two saved projects same basename; non-TTY → error listing candidates |
| 113 | projects/basename-fallback/other-mode/no-fallback | `wrk myrepo --list` with saved project → no fallback, `does not exist` |
| 114 | dep/basename-fallback/single-match/basic | Saved dep; consumer requires module; `wrk --dep mydep` → external wt from saved path |
| 115 | dep/basename-fallback/cwd-exists/no-fallback | `./mydep` in consumer cwd (non-git); saved dep exists → local path, `not a git repository` |
| 116 | dep/basename-fallback/path-with-separator/no-fallback | `wrk --dep sub/mydep` missing → no fallback, `does not exist` |
| 117 | dep/basename-fallback/no-match/error | No saved dep, no local path → `does not exist` |
| 118 | dep/basename-fallback/ambiguous/tty-select | Two saved deps same basename; TTY + stdin selects one; `--dep` succeeds |
| 119 | dep/basename-fallback/ambiguous/non-tty | Two saved deps same basename; non-TTY → error listing candidates |

## How to Run

```sh
# Verify tree structure (no test execution)
doctest vet ./tests

# Fast discovery run (skips labeled leaves — slow perf/multi-repo fixtures)
doctest test ./tests

# Slow / perf leaves only (7 leaves: 12-worktree perf, multi-repo --projects, linked list, output-streaming)
doctest test --label slow ./tests

# Full CI: fast suite then slow suite
doctest test ./tests && doctest test --label slow ./tests

# Flaky timing budget (subset of slow)
doctest test --label flaky ./tests/projects/perf-profile/budget/many-worktrees-parallel

# Run a specific leaf
doctest test ./tests/create-worktree/main-checkout/basic-create

# Run a done leaf
doctest test ./tests/done/ahead-confirm

# Run a list leaf
doctest test ./tests/list/main-only

# Run status leaves
doctest test ./tests/status
doctest test ./tests/status/valid-git-cwd/dirty-counts
doctest vet ./tests/status/master-field
doctest test ./tests/status/master-field
doctest vet ./tests/status/color-output
doctest test ./tests/status/color-output

# Run main-repo-worktrees append leaves (expect RED until append phase implemented)
doctest vet ./tests/status/main-repo-worktrees
doctest test ./tests/status/main-repo-worktrees
doctest test ./tests/status/main-repo-worktrees/external-clean

# Run a dep leaf
doctest test ./tests/dep/basic

# Run --dep basename-fallback leaves (expect RED until resolveDirArg wired in runDep)
doctest vet ./tests/dep/basename-fallback
doctest test ./tests/dep/basename-fallback
doctest test ./tests/dep/basename-fallback/single-match/basic
doctest test ./tests/dep/basename-fallback/ambiguous/tty-select

# Run an all-deps leaf
doctest test ./tests/all-deps/registered/basic

# Run a dry-run leaf
doctest test ./tests/all-deps/registered/dry-run/basic

# Run yes-flag / cascade guard leaves (expect RED until -y + option A implemented)
doctest vet ./tests/yes-flag
doctest test ./tests/yes-flag
doctest test ./tests/done/cascade-force-removal
doctest test ./tests/done/cascade-non-tty-rejects-with-confirm-from-stdin

# Run a done cascade leaf
doctest test ./tests/done/external-cascade
doctest test ./tests/done/cascade-non-external-linked
doctest test ./tests/done/cascade-external-and-deps

# Run a local-replace guard leaf
doctest test ./tests/done/local-replace-blocks
doctest test ./tests/done/sub-module-replace-blocks

# Run an intra-replace (lenient/strict) leaf
doctest test ./tests/done/intra-replace-warns
doctest test ./tests/done/intra-replace-cross-worktree
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
doctest test ./tests/task/spawn/t-alias
doctest test ./tests/task/spawn/empty-task
doctest test ./tests/task/spawn/sequence

# Run a task set-task leaf (non-TTY, expects error)
doctest test ./tests/task/set-task/non-tty
doctest test ./tests/task/set-task/empty-desc
doctest test ./tests/task/set-task/not-linked

# Run a set-task with-dir leaf
doctest test ./tests/task/set-task/with-dir/rename-succeeds
doctest test ./tests/task/set-task/with-dir/missing-dir

# Run projects leaves (expect RED until project persistence is implemented)
doctest vet ./tests/projects
doctest test ./tests/projects
doctest test ./tests/projects/auto-record/no-dir/main-cwd
doctest test ./tests/projects/list/projects/after-records
doctest vet ./tests/projects/detailed-status
doctest test ./tests/projects/detailed-status
doctest vet ./tests/projects/remote-brief
doctest test ./tests/projects/remote-brief
doctest vet ./tests/projects/color-output
doctest test ./tests/projects/color-output
doctest test ./tests/projects/add/manual/main-repo
doctest vet ./tests/projects/remove
doctest test ./tests/projects/remove
doctest test ./tests/projects/remove/manual/main-repo
doctest test ./tests/projects/remove/idempotent/already-removed
doctest test ./tests/projects/events/append-on-success

# Run basename-fallback leaves (expect RED until basename fallback is implemented)
doctest vet ./tests/projects/basename-fallback
doctest test ./tests/projects/basename-fallback
doctest test ./tests/projects/basename-fallback/single-match/create
doctest test ./tests/projects/basename-fallback/ambiguous/tty-select
```

```go
import (
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
	DepPath          string // dep tests: path-to-repo argument
	ConsumerTop      string // dep tests: consumer git toplevel
	ConsumerModDir   string // dep tests: consumer go.mod directory (may differ from repo root for sub-modules)
	ConsumerModDir2  string // dep tests: second consumer go.mod directory for multi-module tests
	ExternalWtDir    string // dep/done tests: external worktree path
	DepsLinkedWtDir  string // done tests: manual linked worktree under deps/ (or other non-external path)
	DepsDepPath      string // done tests: dep main repo that owns DepsLinkedWtDir
	DepModulePath    string // dep tests: module path from dep go.mod
	TaskDesc           string // task tests: task description passed to --task or -t
	TaskFlag           string // task tests: flag form for TaskDesc ("-t" or "--task"; default "--task")
	SetTaskDesc        string // task tests: new task description for --set-task
	SetTaskEnv         string // task tests: extra env vars for --set-task (e.g., WRK_SET_TASK_CONFIRM=1)
	OldExternalGitdir  string // propagate tests: old gitdir content before rename
	ExternalWtDir2    string // propagate tests: second external worktree path
	SecondRepo         string // projects tests: second main repo path
	BasenameEnv        string // basename-fallback tests: e.g. WRK_BASENAME_CONFIRM=1
	SelectedSavedRepo  string // basename-fallback tty-select: chosen saved project path
	ProjectsPerfLog    string // perf-profile tests: WRK_PROJECTS_PERF_LOG path
	FakeHome           string // git-lfs-hook tests: temp home with .local/bin/git-lfs
	UseMinimalPath     bool   // git-lfs-hook tests: run wrk with PATH=/usr/bin:/bin
	UseScriptTTY       bool   // yes-flag tests: run wrk under `script` fake TTY (darwin/linux)
}

type Response struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getWrkBin(t)
	args := buildWrkCLIArgs(req)

	if req.UseScriptTTY {
		return execScriptTTYWrk(t, req, bin, args)
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	cmd.Env = wrkEnv(req)
	return captureCommandOutput(cmd, req.StdinInput)
}
```
