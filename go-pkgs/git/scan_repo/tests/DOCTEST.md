# scan_repo — Git Repository Discovery and Enrichment

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo`. `Scan` walks
filesystem roots to discover git checkouts (`.git` directory or gitlink file).
`ParseRemoteOwnerRepo` parses remote URLs into host, owner, and repo name.
Optional enrichment lists remotes and worktrees via git subprocesses.
Per-directory mirror cache stores `CacheEntry` under a temp or default cache root
via `MirrorEntryPath` / `SaveCacheEntry` / `LoadCacheEntry` (P1 pure store).
Cold `Scan` with `CacheRoot` set and `NoCache=false` populates mirror entries as a
side effect (P2). Warm `Scan` (P3) serves from a complete root cache without a full
re-walk, with liveness checks so deleted repos are never emitted. Warm budgeted
refresh (P4) spends up to `WarmRefreshBudget` rewalking oldest eligible units so
new repos eventually appear. `Options.Refresh` (P5) forces a cold full walk +
rewrite even when warm-eligible. Orphan mirror GC (P7) removes dead child mirror
subtrees when a parent directory is rewalked (cold force or unit refresh).

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies one or more filesystem roots and scan options.
- **Scan** — validates roots, walks each tree, discovers repos, optionally enriches;
  when cache is enabled, writes mirror `CacheEntry` files for visited directories.
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
- **Cache store** — maps a real directory path to a mirror file under
  `<CacheRoot>/mirror/<abs-without-leading-slash>/entry.json`; loads and saves
  `CacheEntry` JSON with atomic rename on write. Tests always pass an explicit
  temp `CacheRoot` (never the product default under `$HOME/.cache/git-repo-scan`).
- **Cold-scan writer** — during a full walk with `NoCache=false` and `CacheRoot`
  set, marks each visited directory in the mirror: repos get `is_repo` /
  `repo_type` / `git_dir`; intermediates get `children` and `scan_complete`.
- **Warm-scan server** — when `NoCache=false`, `CacheRoot` set, and the scan root
  has a usable mirror entry (`scan_complete=true` from a prior cold scan; empty
  `options_hash` treats as match), discovers repos from cached `is_repo` paths
  instead of a full live WalkDir. Verifies each candidate still has a real
  `path/.git` (dir or gitlink); if gone, omits from `Result.Repos` and clears
  the cache mark (`is_repo=false` or remove entry). Missing/incomplete root cache
  falls back to cold full walk + write (P2).
- **Warm budgeted refresher (P4)** — after warm serve, spends up to
  `WarmRefreshBudget` (default **1s** when Options field is 0; **negative** =
  no refresh work) rewalking **refresh units** under the root. A refresh unit
  is each **direct child directory** of the scan root (mirror entry and/or on
  disk). Candidates: `now - refreshed_at >= YoungAge` (default **60s** when 0).
  Sort oldest `refreshed_at` first; skip young units. Rewalk unit = cold-like
  live walk of that child subtree + SaveCacheEntry updates + stamp
  `refreshed_at=now` on visited entries. Merge newly found repos into Result.
  Cold scans remain unlimited full walk (no budget). Optional `Now` clock for
  deterministic tests (nil → `time.Now`).
- **Force refresh (P5)** — when `Options.Refresh=true`, Scan skips the warm path
  and performs a cold full walk + mirror rewrite under `CacheRoot` (unless
  `NoCache`). Unlike budgeted warm refresh, force refresh is unlimited and finds
  brand-new repos warm would miss. Empty `CacheRoot` with cache enabled resolves
  to the product default `$HOME/.cache/git-repo-scan`.
- **Orphan mirror GC (P7)** — when a parent directory is rewalked (cold full walk
  / `Refresh=true`, or budgeted unit rewalk), basenames that existed in the prior
  mirror `children` (or as mirror subdirs) but no longer exist on the live
  filesystem are pruned from the mirror: remove that child's `entry.json` /
  mirror directory subtree so the cache does not grow forever with dead paths.
  Warm serve liveness alone (P3) may only clear `is_repo`; GC is the stronger
  prune that runs on parent rewrite.

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

**Scan (cold cache write — P2)**

- `NoCache=true` — full walk; **no** mirror files written under `CacheRoot`.
- `NoCache=false` and `CacheRoot` set — full cold walk; write `entry.json` for
  visited directories (repos and intermediate dirs); `Result.Repos` unchanged
  vs discovery-only Scan (sort, types, paths).
- Repo directory entry: `is_repo=true`, `repo_type` `"main"`|`"worktree"`,
  `git_dir` absolute (same as `Repo.GitDir`), `refreshed_at` non-empty RFC3339,
  `scan_complete=true`, `version=1`.
- Non-repo intermediate: `is_repo=false`, `children` = immediate child directory
  basenames considered, `scan_complete=true`, `refreshed_at` set.

**Scan (warm serve + liveness — P3)**

- Warm-eligible when `NoCache=false`, `CacheRoot` set, and root mirror entry
  exists with `scan_complete=true` (empty `options_hash` = match).
- Serve candidate repos from cache (`is_repo` entries under root), **not** full
  live WalkDir of the whole tree.
- Liveness: require real `path/.git` still present; else omit from results and
  clear cache mark for that path.
- Soft incompleteness without refresh: brand-new repos under young units or
  with zero refresh budget may be omitted (P3 leaves; P4 budgeted refresh).
- `NoCache=true` always full live walk (finds brand-new; no warm read).
- `Refresh=true` always full cold walk + rewrite (finds brand-new; rewrites
  mirror when cache enabled) even when warm-eligible.
- Incomplete/missing root cache → cold walk + write (P2 behavior).

**Scan (warm budgeted refresh — P4)**

- After warm serve on an eligible root, select refresh units (direct children).
- Eligible when `now - unit.refreshed_at >= YoungAge` (0 YoungAge → default 60s).
- Order oldest first; rewalk until `WarmRefreshBudget` exhausted (0 → default 1s;
  negative → zero refresh work / no unit rewalk).
- Rewalk merges new/changed repos into Result and updates mirror entries.
- Young units are never selected even if budget remains.
- Cold path ignores budget (unlimited full walk).
- Liveness still holds during/after budgeted warm (deleted never emitted).

**Scan (orphan mirror GC — P7)**

- On parent rewalk (cold / `Refresh=true` / unit refresh rewalk), drop mirror
  subtrees for child basenames that no longer exist on disk.
- Acceptable prune: remove `entry.json` (and preferably the mirror directory) for
  the orphan path; parent `children` must not list the dead basename.
- Surviving siblings keep their mirror entries and still appear in `Result`.
- Distinct from P3 liveness: warm-only may leave `is_repo=false` without full
  remove; GC requires entry absent after parent rewrite.

**Enrichment**

- `ListRemotes=false` — no git calls; `Remotes` empty.
- `ListRemotes=true` — `git remote` + config URL per remote on every row.
- `ListWorktrees=false` — no git calls; `Worktrees` empty.
- `ListWorktrees=true` — `git worktree list --porcelain` only on `RepoTypeMain`.

**ParseRemoteOwnerRepo**

- Parse GitHub HTTPS, SSH, and SCP-style URLs into owner and repo.
- Return `ok=false` for unparseable input.

**Cache store (P1)**

- `MirrorEntryPath(cacheRoot, realPath)` — Abs+Clean `realPath`, strip the
  leading path separator, return
  `filepath.Join(cacheRoot, "mirror", <rel-segments>..., "entry.json")`.
- `SaveCacheEntry` — create intermediate mirror dirs; write via temp file then
  rename to `entry.json` (atomic last-writer-wins).
- `LoadCacheEntry` — missing file → `(zero, false, nil)`; corrupt JSON → error;
  valid JSON → `(entry, true, nil)` with field round-trip fidelity.

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
└── cache/                     [explicit temp CacheRoot; P1–P7]
    ├── mirror-path/           [CacheOp=mirror-path]
    │   ├── absolute/          # abs path → mirror/<no-leading-slash>/entry.json
    │   ├── nested/            # multi-segment → nested mirror path segments
    │   └── relative/          # relative Abs-normalized same as abs form
    ├── load/                  [CacheOp=load]
    │   ├── missing/           # no file → ok=false, err=nil
    │   └── corrupt/           # invalid JSON → error
    ├── save/                  [CacheOp=save-load|overwrite]
    │   ├── round-trip/        # Save then Load; all fields preserved
    │   └── overwrite/         # Save A then B; Load returns B; valid JSON on disk
    ├── cold-scan/             [CacheOp empty — Scan with CacheRoot side effects]
    │   ├── write/             [NoCache=false — cold walk writes mirror]
    │   │   ├── main-repo-entry/       # main checkout → is_repo entry
    │   │   ├── intermediate-dirs/     # root/parents → scan_complete + children
    │   │   ├── sibling-repos/         # two mains in cache + Result len 2
    │   │   ├── worktree-entry/        # gitlink → repo_type worktree in cache
    │   │   └── discovery-unchanged/   # Result.Repos same as discovery-only
    │   └── no-cache/          [NoCache=true — no mirror write]
    │       └── skips-write/   # Scan OK; no entry.json under CacheRoot/mirror
    ├── warm/                  [CacheOp empty — second Scan after cold seed; no aging]
    │   ├── serves-cached-omits-new/  # warm: known-repo only; brand-new omitted
    │   ├── omits-deleted/            # warm: gone-repo omitted; cache mark cleared
    │   └── no-cache-finds-new/       # NoCache=true finds brand-new after same seed
    ├── warm-refresh/          [P4 — budgeted unit refresh on warm; stamped times]
    │   ├── discovers-new/     # aged unit + budget → plant under unit found + cached
    │   ├── young-skipped/     # unit within YoungAge → new still omitted
    │   ├── oldest-first/      # two units; tiny budget → only older unit’s new
    │   ├── budget-zero/       # negative WarmRefreshBudget → no refresh; miss new
    │   ├── liveness-holds/    # budgeted warm still omits deleted + clears mark
    │   └── cold-still-full/   # empty cache ignores budget; full discover
    ├── force-refresh/         [P5 nested DOCTEST.md — Options.Refresh=true]
    │   └── finds-new/         # force cold finds brand-new warm would miss
    └── orphan-gc/             [P7 — parent rewalk prunes dead child mirror]
        ├── cold-rescan/       # Refresh=true cold rewalk; gone entry removed
        └── unit-refresh/      # budgeted unit rewalk; orphan under unit removed
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
| `cache/mirror-path/absolute` | Cache | Abs path maps under `mirror/` without leading-slash segment |
| `cache/mirror-path/nested` | Cache | Multi-segment real path → nested mirror path segments |
| `cache/mirror-path/relative` | Cache | Relative real path Abs-normalized same entry as abs form |
| `cache/load/missing` | Cache | Load missing → ok=false, no error |
| `cache/load/corrupt` | Cache | Load invalid JSON → error |
| `cache/save/round-trip` | Cache | Save then Load preserves all CacheEntry fields |
| `cache/save/overwrite` | Cache | Sequential Saves; Load returns last writer; file is valid JSON |
| `cache/cold-scan/write/main-repo-entry` | Cold | Scan writes main repo mirror entry (`is_repo`, `git_dir`, …) |
| `cache/cold-scan/write/intermediate-dirs` | Cold | Scan writes non-repo root/parent entries with `children` |
| `cache/cold-scan/write/sibling-repos` | Cold | Two mains → two cache entries; `Result.Repos` len 2 sorted |
| `cache/cold-scan/write/worktree-entry` | Cold | Gitlink checkout marked `repo_type=worktree` in cache |
| `cache/cold-scan/write/discovery-unchanged` | Cold | Discovery `Repo` shape unchanged when cache writes enabled |
| `cache/cold-scan/no-cache/skips-write` | Cold | `NoCache=true` → no `entry.json` under mirror |
| `cache/warm/serves-cached-omits-new` | Warm | After cold seed, warm returns known-repo; omits planted brand-new |
| `cache/warm/omits-deleted` | Warm | After cold seed + delete dir, warm omits gone-repo; cache mark cleared |
| `cache/warm/no-cache-finds-new` | Warm | `NoCache=true` full live finds brand-new that warm would miss |
| `cache/warm-refresh/discovers-new` | WarmRefresh | Age unit past YoungAge; enough budget → new under unit in Result + cache |
| `cache/warm-refresh/young-skipped` | WarmRefresh | Unit within YoungAge; large budget still omits new under unit |
| `cache/warm-refresh/oldest-first` | WarmRefresh | Two aged units; tiny budget refreshes oldest only → its new found |
| `cache/warm-refresh/budget-zero` | WarmRefresh | Negative WarmRefreshBudget → no unit rewalk; new omitted |
| `cache/warm-refresh/liveness-holds` | WarmRefresh | Budgeted warm still omits deleted repo; cache mark cleared |
| `cache/warm-refresh/cold-still-full` | WarmRefresh | No prior cache; budget options set; cold full walk finds all |
| `cache/force-refresh/finds-new` | ForceRefresh (nested) | P5 `Options.Refresh=true` force cold finds brand-new |
| `cache/orphan-gc/cold-rescan` | OrphanGC | P7 cold `Refresh` rewalk removes orphan `gone` mirror entry |
| `cache/orphan-gc/unit-refresh` | OrphanGC | P7 budgeted unit rewalk removes orphan under unit parent |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/
doctest test -v ./go-pkgs/git/scan_repo/tests/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/
# P5 nested Options.Refresh tree (own DOCTEST.md):
doctest vet ./go-pkgs/git/scan_repo/tests/cache/force-refresh/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/force-refresh/

```

From monorepo root:

```sh
doctest test -v ./external/dot-pkgs-cli/go-pkgs/git/scan_repo/tests/
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

	// Cache store (P1). Non-empty CacheOp dispatches to cache APIs only.
	// Values: "mirror-path" | "load" | "save" | "save-load" | "overwrite".
	// Empty CacheOp with Roots set runs Scan (optionally with CacheRoot / NoCache for P2–P7).
	// Warm / warm-refresh / orphan-gc leaves cold-seed in Setup, then Run performs the second Scan under test.
	// P5 nested force-refresh/ still has its own DOCTEST.md; parent also passes Refresh for P7 cold-rescan.
	CacheOp   string
	CacheRoot string
	NoCache   bool // Scan: when true, do not read or write mirror cache (bypasses warm)
	Refresh   bool // Scan: force cold full walk + mirror rewrite (P5/P7)
	RealPath  string
	Entry     scan_repo.CacheEntry // primary entry for save / save-load / overwrite first write
	EntryB    scan_repo.CacheEntry // second entry for overwrite

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

	// Cache store results
	MirrorPath string
	Entry      scan_repo.CacheEntry
	EntryOK    bool
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.CacheOp != "" {
		switch req.CacheOp {
		case "mirror-path":
			p, err := scan_repo.MirrorEntryPath(req.CacheRoot, req.RealPath)
			if err != nil {
				return nil, err
			}
			return &Response{MirrorPath: p}, nil
		case "load":
			entry, ok, err := scan_repo.LoadCacheEntry(req.CacheRoot, req.RealPath)
			if err != nil {
				return nil, err
			}
			return &Response{Entry: entry, EntryOK: ok}, nil
		case "save":
			if err := scan_repo.SaveCacheEntry(req.CacheRoot, req.RealPath, req.Entry); err != nil {
				return nil, err
			}
			return &Response{}, nil
		case "save-load":
			if err := scan_repo.SaveCacheEntry(req.CacheRoot, req.RealPath, req.Entry); err != nil {
				return nil, err
			}
			entry, ok, err := scan_repo.LoadCacheEntry(req.CacheRoot, req.RealPath)
			if err != nil {
				return nil, err
			}
			return &Response{Entry: entry, EntryOK: ok}, nil
		case "overwrite":
			if err := scan_repo.SaveCacheEntry(req.CacheRoot, req.RealPath, req.Entry); err != nil {
				return nil, err
			}
			if err := scan_repo.SaveCacheEntry(req.CacheRoot, req.RealPath, req.EntryB); err != nil {
				return nil, err
			}
			entry, ok, err := scan_repo.LoadCacheEntry(req.CacheRoot, req.RealPath)
			if err != nil {
				return nil, err
			}
			return &Response{Entry: entry, EntryOK: ok}, nil
		default:
			return nil, fmt.Errorf("unknown CacheOp %q", req.CacheOp)
		}
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
	if err != nil {
		return nil, err
	}
	return &Response{Repos: result.Repos, RootErrors: result.RootErrors}, nil
}
```
