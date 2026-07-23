# scan_repo — Git Repository Discovery and Enrichment

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo`. `Scan` walks
filesystem roots to discover git checkouts (`.git` directory or gitlink file).
`ParseRemoteOwnerRepo` parses remote URLs into host, owner, and repo name.
Optional enrichment lists remotes and worktrees via git subprocesses.

**Dense mirror cache is retired.** v2 durable state is only:
- **repo index** — `home|root/repos.json`
- **walk log** — `home/walk.jsonl` + cursor + meta

Cold `Scan` with `CacheRoot` set and `NoCache=false` seeds the index and walk log
only — it must **not** create `<CacheRoot>/mirror`. Warm `Scan` serves from the
repo index without a full re-walk, with liveness so deleted repos are never
emitted; warm may also **sibling-probe** (`ReadDir` of indexed parents) for
uncached checkouts. Warm budgeted unit refresh spends up to `WarmRefreshBudget`
rewalking oldest eligible units (unit age from live directory ModTime). Separately,
cold Scan appends **walk JSONL** and a cursor; later Scans **consume** under an
adaptive sync budget. `Options.Refresh` forces a cold full walk + rewrite even
when warm-eligible. Nested trees: `cache/repo-index/`, `cache/index-serve/`,
`cache/walk-log/`, `cache/no-mirror/`, `post-filter/` (P1 return-value base-path
filter after optional worktree resolve; classic TDD RED until implementer).

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies one or more filesystem roots and scan options.
- **Scan** — validates roots, walks each tree, discovers repos, optionally enriches;
  when cache is enabled, seeds durable index + walk log (no dense mirror).
- **Walk** — `filepath.WalkDir` from each root; applies ignore config (full paths
  and basenames); on permission errors skips the directory (`SkipDir`) instead of
  aborting; after recording a repo continues into its working tree to discover
  nested checkouts (`.git` basename dirs are still skipped via ignore config).
- **Ignore config** — `IgnoreDirs` are normalized full paths (exact match);
  `IgnoreDirBasenames` union default basenames (`.git`, `node_modules`, …) for
  directory name matches anywhere in the tree.
- **Repo detector** — classifies `.git` as directory (`RepoTypeMain`) or gitlink
  file (`RepoTypeWorktree`); resolves `GitDir` to absolute storage path.
- **Enricher** — when `ListRemotes` or `ListWorktrees` is set, runs `git -C`
  subprocesses per discovered entry; worktrees attach only to main rows.
- **ParseRemoteOwnerRepo** — pure URL parser; no filesystem or git access.
- **Dense mirror (retired)** — historical `<CacheRoot>/mirror/.../entry.json`
  is **not** written or read by Scan. Leaves under `cache/no-mirror/` assert the
  path is absent after cold/warm Scans. Tests always pass an explicit temp
  `CacheRoot` (never the product default under `$HOME/.cache/git-repo-scan`).
- **Repo index store** — durable per-universe index at
  `<CacheRoot>/home/repos.json` and `<CacheRoot>/root/repos.json` (schema v1:
  `version`, `universe`, `base`, `updated_at`, `repos[]` with path/type/git_dir/
  depth/seen_at). `LoadRepoIndex` / `SaveRepoIndex` / `ApplyLiveness`. Nested
  `cache/repo-index/` (own DOCTEST.md) exercises pure I/O; Scan seeds/serves via
  `cache/index-serve/`.
- **Cold-scan writer** — during a full walk with `NoCache=false` and `CacheRoot`
  set, **seeds** the universe `repos.json` and appends **walk log** visits. Does
  **not** write dense mirror entries.
- **Warm-scan / index server** — when `NoCache=false`, `CacheRoot` set, and a
  usable universe `repos.json` has entries under the root, discovers repos from
  the **repo index** instead of a full live WalkDir. Verifies each candidate
  still has a real `path/.git` (dir or gitlink); if gone, omits from
  `Result.Repos` and drops from index. **Sibling probe**: `ReadDir` on a parent
  of an indexed repo can discover a new sibling checkout without full cold
  re-walk. Missing index falls back to cold full walk + write.
- **Walk log store** — append-only JSONL at `<CacheRoot>/home/walk.jsonl` with
  durable cursor `<CacheRoot>/home/walk.cursor.json` and meta
  `home/meta.json` (`last_scan_end`). Events: `visit`, `gone`, `gen_end` (and
  optional repo ops). Nested `cache/walk-log/` (own DOCTEST.md).
- **Walk-log consumer (adaptive budget)** — on warm path, re-lists visit dirs
  from the cursor under `WalkConsumeSyncBudget(Now − last_scan_end)`:
  delta **&lt; 10s** → **0** sync re-list; **10s ≤ delta &lt; 60s** → **500ms**;
  **delta ≥ 60s** → **1s**. When generation `gen_end` G is fully consumed, seal
  `gen_end` **G+1** and advance the cursor to the new EOF.
- **Warm budgeted refresher** — after warm serve, spends up to
  `WarmRefreshBudget` (default **1s** when Options field is 0; **negative** =
  no refresh work) rewalking **refresh units** under the root. A refresh unit
  is each **direct child directory** of the scan root on disk. Candidates:
  `now - unit_dir.ModTime >= YoungAge` (default **60s** when 0). Sort oldest
  first; skip young units. Rewalk unit = cold-like live walk of that child
  subtree + index update. Merge newly found repos into Result. A single unit
  rewalk must also respect remaining budget (child context deadline; soft
  partial merge on expiry). Cold scans remain unlimited full walk (no budget).
  Optional `Now` clock for deterministic tests (nil → `time.Now`). Orthogonal
  to walk-log consume budget.
- **Warm refresh mode (sync default / async opt-in)** —
  `Options.WarmRefreshMode`: `WarmRefreshSync` (zero value) runs unit refresh +
  walk-log consume inside `Scan`/`ScanSession` before return (today).
  `WarmRefreshAsync` (via `ScanSession` only; classic `Scan` forces Sync): after
  warm serve, returns a `Session` with Result frozen at the serve snapshot and
  a `RefreshHandle`. Background polish updates durable index/walk log only
  (no OnRepo / Result mutation). Join rule: keep polishing while work remains
  and (`now < start+budget` **or** Join not yet requested); on Join before
  budget wait until budget or idle; on Join after budget soft-stop; `Stop`
  aborts min-budget wait and keeps already-written index. Idle (no work) →
  Join returns immediately (budget is max effort, not sleep).
- **Force refresh** — when `Options.Refresh=true`, Scan skips the warm path
  and performs a cold full walk + index/walk rewrite under `CacheRoot` (unless
  `NoCache`). Unlike budgeted warm refresh, force refresh is unlimited and finds
  brand-new repos warm would miss. Empty `CacheRoot` with cache enabled resolves
  to the product default `$HOME/.cache/git-repo-scan`.
- **Debug logger** — when `Options.Debug=true`, Scan writes phase-level structured
  `scan:` lines to `Options.Stderr` (default `os.Stderr`): cache root, per-root
  `mode=warm|cold` + reason, warm serve candidate/live counts and duration,
  refresh budget/units summary, root total. When false, zero `scan:` markers
  (orthogonal to `Verbose` permission/remote skip warnings). Nested
  `cache/debug/` (own DOCTEST.md) exercises Debug on/off with stderr capture.

### Behaviors

**Scan (discovery)**

- Require at least one root; invalid roots (missing path, not a directory) are
  recorded in `RootErrors` and scanning continues for remaining roots.
- Expand `~`, absolutize and clean paths; sort results by `Path` ascending.
- Apply default ignore basenames unioned with `IgnoreDirBasenames`.
- Skip directories whose normalized full path is listed in `IgnoreDirs`.
- When `Verbose` is true, log permission-denied and remote-backed filesystem
  skips to stderr as warnings.
- Skip `Library/CloudStorage/...` paths before walking them (macOS cloud-sync roots).
- Respect `MaxDepth` relative to each root (0 = unlimited).
- Option A: every checkout with `.git` is its own row; no dedup by `GitDir`.

**Scan (cold cache write)**

- `NoCache=true` — full walk; **no** `repos.json` seed, **no** walk log, **no**
  `mirror/` under `CacheRoot`.
- `NoCache=false` and `CacheRoot` set — full cold walk; seed universe
  `repos.json`; append walk visits and seal `gen_end` gen=1 with cursor at EOF;
  `Result.Repos` unchanged vs discovery-only Scan (sort, types, paths).
- **Must not** create `<CacheRoot>/mirror` (dense mirror retired).
- Index repo rows: path, `repo_type` `"main"`|`"worktree"`, `git_dir` absolute,
  `seen_at` non-empty RFC3339, schema `version=1`.

**Scan (warm serve + index + liveness + sibling)**

- Warm-eligible when `NoCache=false`, `CacheRoot` set, and a usable universe
  `repos.json` has entries under the scan root.
- Serve candidate repos from the **repo index**, **not** full live WalkDir of
  the whole tree (no mirror `is_repo` marks).
- Liveness: require real `path/.git` still present; else omit from results and
  drop dead index entries.
- Sibling: after cold indexes `parent/A`, planting `parent/B/.git` is found on
  warm Scan via parent `ReadDir` without `Refresh=true`.
- Soft incompleteness without refresh/consume: brand-new repos outside sibling
  reach under young units or with zero budgets may be omitted.
- `NoCache=true` always full live walk (finds brand-new; no warm read).
- `Refresh=true` always full cold walk + rewrite (finds brand-new; rewrites
  index/walk when cache enabled) even when warm-eligible.
- Missing index → cold walk + write.

**Scan (walk log cold seal + adaptive consume)**

- Cold success with cache: `home/walk.jsonl` has visit lines; ends with
  `{"op":"gen_end","gen":1}`; `home/walk.cursor.json` offset = sealed EOF.
- Warm consume: from cursor, re-list prior visits under adaptive
  `WalkConsumeSyncBudget`; may append `gone` / new visits; when `gen_end` G is
  consumed, append `gen_end` G+1 and advance cursor to new EOF.
- Budget tiers (exported `WalkConsumeSyncBudget`): &lt;10s → 0; 10–60s → 500ms;
  ≥60s → 1s. Tests inject `Options.LastScanEnd` + `Options.Now` (or meta).
- Consume path often disables unit warm-refresh (`WarmRefreshBudget=-1`) in
  leaves so observations isolate walk-log work.

**Scan (warm budgeted unit refresh)**

- After warm serve on an eligible root, select refresh units (direct children on disk).
- Eligible when `now - unit_dir.ModTime >= YoungAge` (0 YoungAge → default 60s).
- Order oldest first; rewalk until `WarmRefreshBudget` exhausted (0 → default 1s;
  negative → zero refresh work / no unit rewalk).
- Budget bounds **within** a single unit rewalk as well as between units: unit
  walk uses a **child context with remaining-budget deadline**; parent Scan /
  SIGINT context is not cancelled. Mid-unit budget expiry is soft.
- Rewalk merges new/changed repos into Result and updates `home/repos.json`.
- Young units are never selected even if budget remains.
- Cold path ignores budget (unlimited full walk).
- Liveness still holds during/after budgeted warm (deleted never emitted).
- Dense mirror is not written during rewalk.

**Scan (debug phase logs — `Options.Debug`)**

- `Debug=true` → greppable `scan:` phase lines on `Options.Stderr` (default
  `os.Stderr`): cache root, `mode=warm|cold` + reason, warm serve timing,
  refresh budget summary, root total. Volume stays phase-level (not one line
  per visited directory).
- `Debug=false` → zero `scan:` markers; `Verbose` skip warnings keep their own
  format and are not required to carry the `scan:` prefix.
- Exercised by nested `cache/debug/` leaves (`on/cold`, `on/warm`, `off`).

**Repo index pure I/O (nested `cache/repo-index/`)**

- Save/Load round-trip under universe `home` and `root`; load missing → empty,
  not error; `ApplyLiveness` drops dead `.git` paths.

**Enrichment**

- `ListRemotes=false` — no git calls; `Remotes` empty.
- `ListRemotes=true` — `git remote` + config URL per remote on every row.
- `ListWorktrees=false` — no git calls; `Worktrees` empty.
- `ListWorktrees=true` — `git worktree list --porcelain` only on `RepoTypeMain`.

**ParseRemoteOwnerRepo**

- Parse GitHub HTTPS, SSH, and SCP-style URLs into owner and repo.
- Return `ok=false` for unparseable input.

**Cache store (index + walk helpers; mirror retired)**

- Dense mirror APIs (`MirrorEntryPath` / `SaveCacheEntry` / `LoadCacheEntry`) are
  **not** part of the Scan contract and must not be used by product Scan paths.
- `SaveRepoIndex` / `LoadRepoIndex` — atomic `home|root/repos.json`.
- `WalkLogAppend` / `WalkConsumeSyncBudget` / seal + cursor helpers — walk JSONL.

## Decision Tree

```
scan-repo
├── parse-remote/              [req.ParseURL set — pure parser, no FS/git]
│   ├── parse-github-ssh/
│   ├── parse-github-https/
│   ├── parse-scp-style/
│   └── parse-invalid/
├── scan/                      [ListRemotes=false, ListWorktrees=false]
│   ├── single-repo/
│   ├── sibling-repos/
│   ├── no-repos/
│   ├── repo-boundary/
│   ├── max-depth/
│   ├── ignore-dirs/              [default basename: node_modules]
│   ├── ignore-dir-basename/      [custom IgnoreDirBasenames]
│   ├── ignore-dir-full-path/     [IgnoreDirs full path]
│   ├── permission-denied-skip/   [WalkDir EACCES → SkipDir]
│   ├── skip-cloud-storage/       [CloudStorage subtree → SkipDir]
│   ├── remote-root-skip/         [CloudStorage root → empty result]
│   ├── gitlink-worktree/
│   ├── main-and-linked/
│   ├── nested-under-checkout/    [scan root IS a checkout — nested repos inside must be found]
│   │   ├── external-linked/      # wt root + external/mydep linked wt → 2 rows
│   │   └── nested-main/          # wt root + vendor/nested main repo → 2 rows
│   ├── empty-roots-error/
│   ├── missing-root-error/          # RootError; scan err nil
│   ├── not-a-directory-error/       # RootError; scan err nil
│   └── root-failure-isolated/       # good root + bad root → partial result
├── enrich-remotes/            [ListRemotes=true, ListWorktrees=false]
│   ├── no-remotes/
│   ├── single-origin/
│   ├── multiple-remotes/
│   └── flags-false-skips-git/
├── enrich-worktrees/          [ListWorktrees=true]
│   ├── main-only/
│   ├── main-plus-linked/
│   └── flags-false-skips-git/
├── find-github/               [FindLocalMainByGitHub]
│   ├── basename-mismatch/     clone dir name != github repo name
│   └── skips-worktree/        returns main, not linked worktree
├── cache/                     [explicit temp CacheRoot; index + walk; mirror retired]
    ├── no-mirror/             [RED until product stops writing mirror/]
    │   ├── cold-scan-no-mirror-dir/  # cold Scan → no CacheRoot/mirror
    │   └── warm-no-mirror-growth/    # after seed + warm → still no mirror/
    ├── cold-scan/             [CacheOp empty — Scan with CacheRoot side effects]
    │   ├── write/             [NoCache=false — cold walk seeds index/walk]
    │   │   ├── main-repo-in-index/    # main checkout in home/repos.json
    │   │   ├── sibling-repos-in-index/# two mains in index + Result len 2
    │   │   ├── worktree-in-index/     # gitlink → repo_type worktree in index
    │   │   └── discovery-unchanged/   # Result.Repos same as discovery-only
    │   └── no-cache/          [NoCache=true — no durable cache write]
    │       └── skips-write/   # Scan OK; no home/repos.json, walk, or mirror
    ├── warm/                  [second Scan after cold seed; soft-omit; index serve]
    │   ├── serves-cached-omits-new/  # warm: known-repo only; brand-new omitted
    │   ├── omits-deleted/            # warm: gone-repo omitted; index drop
    │   └── no-cache-finds-new/       # NoCache=true finds brand-new after same seed
    ├── warm-refresh/          [P4 — budgeted unit refresh; unit ModTime stamps]
    │   ├── discovers-new/     # aged unit + budget → plant under unit found + indexed
    │   ├── young-skipped/     # unit within YoungAge → new still omitted
    │   ├── oldest-first/      # two units; tiny budget → only older unit’s new
    │   ├── budget-zero/       # negative WarmRefreshBudget → no refresh; miss new
    │   ├── budget-caps-unit-walk/ # tiny budget + huge unit → Scan finishes fast; seed kept
    │   ├── liveness-holds/    # budgeted warm still omits deleted + index drop
    │   └── cold-still-full/   # empty cache ignores budget; full discover
    ├── force-refresh/         [P5 nested DOCTEST.md — Options.Refresh=true]
    │   └── finds-new/         # force cold finds brand-new warm would miss
    ├── debug/                 [nested DOCTEST.md — Options.Debug scan: timing/mode]
    │   ├── on/cold/           # Debug=true empty cache → mode=cold + scan:
    │   ├── on/warm/           # Debug=true after seed → mode=warm + serve timing
    │   └── off/               # Debug=false → zero scan: markers
    ├── repo-index/            [nested — Load/SaveRepoIndex + ApplyLiveness]
    │   ├── save-load/home/    # universe=home → CacheRoot/home/repos.json
    │   ├── save-load/root/    # universe=root → CacheRoot/root/repos.json
    │   ├── load/missing/      # no repos.json → empty / ok=false
    │   └── liveness/drop-dead/ # ApplyLiveness drops missing-.git
    ├── index-serve/           [nested — Scan seeds/serves home/repos.json + sibling]
    │   ├── cold-seed/writes-index/       # cold writes home/repos.json mains
    │   ├── warm-serve/from-index/        # warm Result from index + IndexOK
    │   ├── sibling/discovers-new/        # uncached sibling via ReadDir
    │   └── liveness/drop-dead-via-scan/  # dead indexed path omitted on Scan
    └── walk-log/              [nested — walk.jsonl + gen_end + adaptive budget]
        ├── cold/complete/     # visits + gen_end gen=1 + cursor at EOF
        ├── no-cache/skips-write/  # NoCache → no walk artifacts under home/
        └── consume/           # second Scan consume from cursor
            ├── seal-gen-end-2/       # process gen_end 1 → seal gen_end 2
            ├── cursor-advance/       # cursor > cold EOF; equals new size
            ├── gone/                 # deleted visit path → gone event
            └── budget/               # WalkConsumeSyncBudget tiers
                ├── delta-lt-10s/     # 0 sync; no consume discover
                ├── delta-10s-to-60s/ # 500ms
                └── delta-ge-60s/     # 1s
└── post-filter/               [nested — P1 resolve then base-path filter; RED until implementer]
    ├── walk-log-foreign-leak/                 # warm consume must not emit foreign agent-pro
    ├── list-worktrees-inner-only/             # under-root wt on main.Worktrees
    ├── list-worktrees-outside-base-stripped/  # outer wt stripped from Worktrees
    └── list-worktrees-false-top-level-filter/ # flag off: Worktrees empty + top-level filter
```

## Test Index

| Leaf | Mode | Description |
|------|------|-------------|
| `parse-remote/parse-github-ssh` | Parse | SSH URL → owner/repo |
| `parse-remote/parse-github-https` | Parse | HTTPS URL → owner/repo |
| `parse-remote/parse-scp-style` | Parse | Enterprise SCP URL → owner/repo |
| `parse-remote/parse-invalid` | Parse | Unparseable URL → ok=false |
| `scan/single-repo` | Scan | One main repo discovered |
| `scan/sibling-repos` | Scan | Two sibling repos, path-sorted |
| `scan/no-repos` | Scan | Empty tree → empty slice |
| `scan/repo-boundary` | Scan | Nested `.git` inside found repo discovered |
| `scan/max-depth` | Scan | Deep repo excluded by MaxDepth |
| `scan/ignore-dirs` | Scan | `node_modules` default basename ignore |
| `scan/ignore-dir-basename` | Scan | Custom `IgnoreDirBasenames` skips tree |
| `scan/ignore-dir-full-path` | Scan | `IgnoreDirs` exact full path skips tree |
| `scan/permission-denied-skip` | Scan | Unreadable child dir; scan still succeeds |
| `scan/skip-cloud-storage` | Scan | CloudStorage subtree skipped; local repo found |
| `scan/remote-root-skip` | Scan | CloudStorage scan root yields no repos |
| `scan/gitlink-worktree` | Scan | Gitlink → RepoTypeWorktree |
| `scan/main-and-linked` | Scan | Main + linked worktree as two rows |
| `scan/nested-under-checkout/external-linked` | Scan | Scan from wt root finds nested `external/mydep` linked wt |
| `scan/nested-under-checkout/nested-main` | Scan | Scan from wt root finds nested `vendor/nested` main repo |
| `scan/empty-roots-error` | Scan | No roots → error |
| `scan/missing-root-error` | Scan | Missing root path → RootError; err nil |
| `scan/not-a-directory-error` | Scan | File root → RootError; err nil |
| `scan/root-failure-isolated` | Scan | Valid repo + missing root → 1 repo + 1 RootError |
| `enrich-remotes/no-remotes` | Enrich | Git init, empty Remotes |
| `enrich-remotes/single-origin` | Enrich | Single origin remote parsed |
| `enrich-remotes/multiple-remotes` | Enrich | origin + upstream remotes |
| `enrich-remotes/flags-false-skips-git` | Enrich | ListRemotes=false, fake repo OK |
| `enrich-worktrees/main-only` | Enrich | Worktrees on main row only |
| `enrich-worktrees/main-plus-linked` | Enrich | Two rows; Worktrees only on main |
| `enrich-worktrees/flags-false-skips-git` | Enrich | ListWorktrees=false, fake repo OK |
| `find-github/basename-mismatch` | Find | `myproject-clone` + origin `xhd2015/myproject` |
| `find-github/skips-worktree` | Find | linked worktree skipped; main path returned |
| `cache/no-mirror/cold-scan-no-mirror-dir` | NoMirror (RED) | Cold Scan must not create `<CacheRoot>/mirror` |
| `cache/no-mirror/warm-no-mirror-growth` | NoMirror (RED) | Warm path must not leave/create `mirror/` |
| `cache/cold-scan/write/main-repo-in-index` | Cold | Scan seeds main in `home/repos.json`; no mirror |
| `cache/cold-scan/write/sibling-repos-in-index` | Cold | Two mains → two index entries; Result len 2 |
| `cache/cold-scan/write/worktree-in-index` | Cold | Gitlink marked `repo_type=worktree` in index |
| `cache/cold-scan/write/discovery-unchanged` | Cold | Discovery `Repo` shape unchanged when cache writes enabled |
| `cache/cold-scan/no-cache/skips-write` | Cold | `NoCache=true` → no index / walk / mirror |
| `cache/warm/serves-cached-omits-new` | Warm | After cold seed, warm returns known-repo; omits planted brand-new |
| `cache/warm/omits-deleted` | Warm | After cold seed + delete dir, warm omits gone-repo; index drop |
| `cache/warm/no-cache-finds-new` | Warm | `NoCache=true` full live finds brand-new that warm would miss |
| `cache/warm-refresh/discovers-new` | WarmRefresh | Age unit past YoungAge; enough budget → new under unit in Result + index |
| `cache/warm-refresh/young-skipped` | WarmRefresh | Unit within YoungAge; large budget still omits new under unit |
| `cache/warm-refresh/oldest-first` | WarmRefresh | Two aged units; tiny budget refreshes oldest only → its new found |
| `cache/warm-refresh/budget-zero` | WarmRefresh | Negative WarmRefreshBudget → no unit rewalk; new omitted |
| `cache/warm-refresh/budget-caps-unit-walk` | WarmRefresh | Tiny budget + huge eligible unit → Scan wall << unbounded rewalk; known seed still served |
| `cache/warm-refresh/liveness-holds` | WarmRefresh | Budgeted warm still omits deleted repo; index drop |
| `cache/warm-refresh/cold-still-full` | WarmRefresh | No prior cache; budget options set; cold full walk finds all |
| `cache/force-refresh/finds-new` | ForceRefresh (nested) | P5 `Options.Refresh=true` force cold finds brand-new |
| `cache/debug/on/cold` | Debug (nested) | `Debug=true` empty cache → stderr `scan:` + `mode=cold` |
| `cache/debug/on/warm` | Debug (nested) | `Debug=true` after cold seed → `mode=warm` + serve timing |
| `cache/debug/off` | Debug (nested) | `Debug=false` → zero `scan:` markers on stderr |
| `cache/repo-index/save-load/home` | RepoIndex (nested) | Round-trip v1 fields under `home/repos.json` |
| `cache/repo-index/save-load/root` | RepoIndex (nested) | Round-trip v1 fields under `root/repos.json` |
| `cache/repo-index/load/missing` | RepoIndex (nested) | Missing file → empty index, not error |
| `cache/repo-index/liveness/drop-dead` | RepoIndex (nested) | `ApplyLiveness` drops dead; keeps live |
| `cache/index-serve/cold-seed/writes-index` | IndexServe (nested) | Cold Scan seeds `home/repos.json` mains |
| `cache/index-serve/warm-serve/from-index` | IndexServe (nested) | Warm serves indexed live mains |
| `cache/index-serve/sibling/discovers-new` | IndexServe (nested) | Sibling ReadDir finds uncached B |
| `cache/index-serve/liveness/drop-dead-via-scan` | IndexServe (nested) | Dead indexed path omitted on Scan |
| `cache/walk-log/cold/complete` | WalkLog (nested) | Cold visits + `gen_end` 1 + cursor EOF |
| `cache/walk-log/no-cache/skips-write` | WalkLog (nested) | NoCache → no walk.jsonl / cursor |
| `cache/walk-log/consume/seal-gen-end-2` | WalkLog (nested) | Consume seals `gen_end` 2 |
| `cache/walk-log/consume/cursor-advance` | WalkLog (nested) | Cursor advances past cold EOF |
| `cache/walk-log/consume/gone` | WalkLog (nested) | Removed visit → `gone` event |
| `cache/walk-log/consume/budget/delta-lt-10s` | WalkLog (nested) | Adaptive budget 0; no consume discover |
| `cache/walk-log/consume/budget/delta-10s-to-60s` | WalkLog (nested) | Budget 500ms tier |
| `cache/walk-log/consume/budget/delta-ge-60s` | WalkLog (nested) | Budget 1s tier |
| `post-filter/walk-log-foreign-leak` | PostFilter (nested, RED) | Warm walk-log consume must not return foreign agent-pro |
| `post-filter/list-worktrees-inner-only` | PostFilter (nested, RED) | Under-root linked wt listed on main.Worktrees |
| `post-filter/list-worktrees-outside-base-stripped` | PostFilter (nested, RED) | Outer worktree stripped from Worktrees after resolve |
| `post-filter/list-worktrees-false-top-level-filter` | PostFilter (nested, RED) | ListWorktrees=false: empty Worktrees + top-level under-root filter |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/
doctest test -v ./go-pkgs/git/scan_repo/tests/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/
# No-mirror contract (RED until product removes mirror writes):
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/no-mirror/
# Nested Options.Refresh tree (own DOCTEST.md):
doctest vet ./go-pkgs/git/scan_repo/tests/cache/force-refresh/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/force-refresh/
# Options.Debug scan: timing/mode logs (own DOCTEST.md):
doctest vet ./go-pkgs/git/scan_repo/tests/cache/debug/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/debug/
# Repo index pure I/O (own DOCTEST.md):
doctest vet ./go-pkgs/git/scan_repo/tests/cache/repo-index/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/repo-index/
# Index seed/serve/sibling/liveness via Scan (own DOCTEST.md):
doctest vet ./go-pkgs/git/scan_repo/tests/cache/index-serve/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/index-serve/
# Walk JSONL cold + gen_end consume + adaptive budget (own DOCTEST.md):
doctest vet ./go-pkgs/git/scan_repo/tests/cache/walk-log/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/walk-log/
# P1 post-process base-path filter (own DOCTEST.md; RED until implementer):
doctest vet ./go-pkgs/git/scan_repo/tests/post-filter/
doctest test -v ./go-pkgs/git/scan_repo/tests/post-filter/
```

From monorepo / worktree root:

```sh
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/repo-index/
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/index-serve/
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/walk-log/
doctest test -v ./external/dot-pkgs-master-2026-07-22/go-pkgs/git/scan_repo/tests/post-filter/
```

```go
import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

type Request struct {
	Roots              []string
	MaxDepth           int
	IgnoreDirs         []string
	IgnoreDirBasenames []string
	Verbose            bool
	ListRemotes        bool
	ListWorktrees      bool
	ParseURL           string // non-empty → ParseRemoteOwnerRepo only
	FindGitHubOwner    string
	FindGitHubRepo     string

	// CacheOp is reserved/empty for Scan leaves (dense mirror pure-store ops retired).
	// Empty CacheOp with Roots set runs Scan (optionally with CacheRoot / NoCache).
	// Warm / warm-refresh leaves cold-seed in Setup, then Run performs the second Scan under test.
	// P5 nested force-refresh/ still has its own DOCTEST.md.
	CacheOp   string
	CacheRoot string
	NoCache   bool // Scan: when true, do not read or write durable cache (bypasses warm)
	Refresh   bool // Scan: force cold full walk + index/walk rewrite
	RealPath  string // leaf scratch (e.g. deleted path abs for liveness asserts)

	// P4 budgeted warm refresh — passed through to scan_repo.Options.
	// WarmRefreshBudget: 0 → product default 1s; negative → no refresh work.
	// YoungAge: 0 → product default 60s.
	// Now: optional clock (nil → time.Now); tests prefer stamped refreshed_at over sleeps.
	WarmRefreshBudget time.Duration
	YoungAge          time.Duration
	Now               func() time.Time
}

type Response struct {
	Repos      []scan_repo.Repo
	RootErrors []scan_repo.RootError
	Found      *scan_repo.Repo
	Owner      string
	Repo       string
	ParseOK    bool

	// Elapsed is wall time of the primary operation in Run (Scan path only).
	// Used by budget-caps leaves to prove WarmRefreshBudget bounds mid-unit work.
	Elapsed time.Duration
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.CacheOp != "" {
		return nil, fmt.Errorf("unknown/retired CacheOp %q (dense mirror store removed)", req.CacheOp)
	}
	if req.ParseURL != "" {
		owner, repo, ok := scan_repo.ParseRemoteOwnerRepo(req.ParseURL)
		return &Response{Owner: owner, Repo: repo, ParseOK: ok}, nil
	}
	if req.FindGitHubOwner != "" || req.FindGitHubRepo != "" {
		found, err := scan_repo.FindLocalMainByGitHub(context.Background(), scan_repo.Options{
			Roots: req.Roots,
		}, req.FindGitHubOwner, req.FindGitHubRepo)
		if err != nil {
			return nil, err
		}
		return &Response{Found: found}, nil
	}
	scanStart := time.Now()
	result, err := scan_repo.Scan(context.Background(), scan_repo.Options{
		Roots:              req.Roots,
		MaxDepth:           req.MaxDepth,
		IgnoreDirs:         req.IgnoreDirs,
		IgnoreDirBasenames: req.IgnoreDirBasenames,
		Verbose:            req.Verbose,
		ListRemotes:        req.ListRemotes,
		ListWorktrees:      req.ListWorktrees,
		CacheRoot:          req.CacheRoot,
		NoCache:            req.NoCache,
		Refresh:            req.Refresh,
		WarmRefreshBudget:  req.WarmRefreshBudget,
		YoungAge:           req.YoungAge,
		Now:                req.Now,
	})
	elapsed := time.Since(scanStart)
	if err != nil {
		return &Response{Elapsed: elapsed}, err
	}
	return &Response{Repos: result.Repos, RootErrors: result.RootErrors, Elapsed: elapsed}, nil
}
```
