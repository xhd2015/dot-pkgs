# ptywrap process env merge (`MergeProcessEnv` + spawn TERM)

Plan phase **P1** contract tests: PTY spawn builds `cmd.Env` by merging a
base environ with caller **set** / **unset**, then applying the spawn **TERM**
policy (default only when final TERM is missing, empty, or `dumb`).

Classic TDD: `SpawnOptions.Env` / `SpawnOptions.Unset` and the pure merge
helpers are **not** implemented yet. This tree is **RED** until the implementer
lands the sealed API below.

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — session/manager path that supplies child env deltas (set/unset)
  when starting a PTY, instead of only inheriting the parent process environ.
- **Base environ** — injectable `[]string` of `KEY=value` entries (in production:
  typically `os.Environ()`; tests pass explicit slices — no process-global env).
- **`MergeProcessEnv`** — pure merge helper: start from base, remove **Unset**
  keys, apply **Set** assignments (`KEY=value`, last-wins).
- **`EnsureSpawnTERM`** — pure TERM policy for the spawn path: if final TERM is
  missing, empty, or exactly `dumb`, set `TERM=xterm-256color`; never clobber a
  good TERM.
- **`SpawnOptions`** — gains `Env []string` and `Unset []string`; spawn wires
  `EnsureSpawnTERM(MergeProcessEnv(base, opts.Env, opts.Unset))` before
  ExtraPaths / PS1 appends (those remain separate).

**Behaviors**

- **Identity** — empty set and unset → result equals base (map of keys; pure
  merge does **not** force TERM).
- **Set** — each `KEY=value` in set is applied after unset; duplicate KEY in set
  → last entry wins; existing base KEY is replaced.
- **Unset** — keys listed in unset are removed from base (all occurrences);
  unsetting a missing key is a no-op.
- **Order of operations** — unset first, then set (so unset-then-set reintroduces
  KEY with the set value).
- **Spawn TERM** — after merge, if TERM missing / `""` / `dumb` →
  `TERM=xterm-256color`; otherwise leave TERM unchanged (including
  `xterm-256color` and other usable values like `screen-256color`).
- **Out of scope for this tree** — ExtraPaths PATH append, PS1, tty-watch,
  agent-pro; document only that ExtraPaths still appends after env merge.

**Product API sealed for implementer**

| Symbol | Role |
|--------|------|
| `MergeProcessEnv(base, set, unset []string) []string` | Pure merge: base → remove unset keys → apply set (last-wins) |
| `EnsureSpawnTERM(env []string) []string` | Spawn TERM policy (missing/empty/`dumb` → `xterm-256color`) |
| `SpawnOptions.Env []string` | Set assignments (`KEY=value`) for child |
| `SpawnOptions.Unset []string` | Keys removed before Env applied |

Spawn wiring (implementer checklist, not asserted here as process spawn):

```text
cmd.Env = EnsureSpawnTERM(MergeProcessEnv(os.Environ(), opts.Env, opts.Unset))
// then existing ExtraPaths / PS1 appends
```

## Version

0.0.2

## Decision Tree

```
shell/ptywrap/tests/process-env/
├── DOCTEST.md
├── SETUP.md
├── merge/                              # pure MergeProcessEnv (no TERM policy)
│   ├── identity/
│   │   └── empty-set-unset/            # S1: equals base
│   ├── set/
│   │   ├── single-assignment/          # S2: KEY present with v
│   │   └── last-wins/                  # S3: same KEY twice → last value
│   ├── unset/
│   │   ├── removes-present/            # S4: KEY absent
│   │   ├── absent-key-noop/            # unset missing key → base unchanged
│   │   └── removes-no-color/           # S8: NO_COLOR absent
│   └── unset-then-set/
│       └── key-reintroduced/           # S5: KEY present with set value
└── spawn-term/                         # merge + EnsureSpawnTERM
    ├── needs-default/
    │   ├── missing-term/               # S6: no TERM → xterm-256color
    │   ├── empty-term/                 # S6: TERM= → xterm-256color
    │   ├── dumb-term/                  # S6: TERM=dumb → xterm-256color
    │   └── after-unset-term/           # unset TERM then policy fills default
    └── preserves-good/
        ├── xterm-256color/             # S7: already default → keep
        └── other-good-term/            # S7: TERM=screen-256color → keep
```

Parameter ranking (most → least significant):

1. **Surface** — pure merge vs spawn TERM policy
2. **Operation class** — identity / set / unset / unset+set, or needs-default vs preserves-good
3. **Concrete inputs** — key values, dumb vs empty vs missing TERM

## Test Index

| # | Leaf | Surface | Description |
|---|------|---------|-------------|
| 1 | `merge/identity/empty-set-unset` | merge | empty set/unset → env equals base |
| 2 | `merge/set/single-assignment` | merge | set `FOO=bar` → FOO=bar present |
| 3 | `merge/set/last-wins` | merge | set FOO twice → last value wins |
| 4 | `merge/unset/removes-present` | merge | unset KEY in base → KEY absent |
| 5 | `merge/unset/absent-key-noop` | merge | unset missing KEY → other keys intact |
| 6 | `merge/unset/removes-no-color` | merge | base has NO_COLOR; unset → absent |
| 7 | `merge/unset-then-set/key-reintroduced` | merge | unset then set KEY → set value |
| 8 | `spawn-term/needs-default/missing-term` | spawn-term | no TERM → TERM=xterm-256color |
| 9 | `spawn-term/needs-default/empty-term` | spawn-term | TERM= → TERM=xterm-256color |
| 10 | `spawn-term/needs-default/dumb-term` | spawn-term | TERM=dumb → TERM=xterm-256color |
| 11 | `spawn-term/needs-default/after-unset-term` | spawn-term | unset TERM → default applied |
| 12 | `spawn-term/preserves-good/xterm-256color` | spawn-term | good default not clobbered |
| 13 | `spawn-term/preserves-good/other-good-term` | spawn-term | TERM=screen-256color kept |

## How to Run

```sh
# from go-pkgs module root (brought tree)
cd go-pkgs   # or the vendored path containing go.mod for this module
doctest vet ./shell/ptywrap/tests/process-env
doctest test ./shell/ptywrap/tests/process-env/...
```

Expect **RED** until implementer exports `MergeProcessEnv` and
`EnsureSpawnTERM` (and wires `SpawnOptions.Env` / `Unset` in spawn).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"
)

// Request drives pure process-env helpers without spawning a PTY.
// Base/Set/Unset are injectable — no t.Setenv / os.Setenv / process environ.
type Request struct {
	Base  []string // KEY=value base environ
	Set   []string // KEY=value assignments (SpawnOptions.Env)
	Unset []string // keys to remove (SpawnOptions.Unset)

	// ApplySpawnTERM when true: EnsureSpawnTERM(MergeProcessEnv(...)).
	// When false: pure MergeProcessEnv only.
	ApplySpawnTERM bool
}

// Response holds the merged environ slice returned by the product API.
type Response struct {
	Env []string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	env := ptywrap.MergeProcessEnv(req.Base, req.Set, req.Unset)
	if req.ApplySpawnTERM {
		env = ptywrap.EnsureSpawnTERM(env)
	}
	return &Response{Env: env}, nil
}
```
