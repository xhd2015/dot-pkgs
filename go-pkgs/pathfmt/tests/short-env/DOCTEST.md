# ShortEnv — `pathfmt.ShortEnv` / `pathfmt.ShortEnvFrom`

## Version
0.0.2

Unit-style doc tests for env-prefix path shortening in
`github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt`. `ShortEnvFrom` replaces the
longest eligible environment path prefix with `$NAME`; `ShortEnv` wraps it
with `os.Environ()`. Display-only — never for I/O.

**Classic TDD:** `ShortEnv` / `ShortEnvFrom` do not exist yet. Root `Run` calls
them so the suite is compile-RED until the implementer adds the API. Do not
substitute `TildeHome` / `Short` for the new symbols.

## DSN (Domain Specific Notion)

### Participants

- **`ShortEnvFrom`** — pure formatter: `(path, env []string) → display`. Builds
  eligible path aliases from `env` (`KEY=value` slices, `os.Environ` style),
  then one longest-prefix replacement at a path-segment boundary. No match →
  `TildeHome(path)`. Empty/`nil` env → `TildeHome` only (no `os.Environ`
  magic). Not cwd-relative (never uses `Short` / `ShortFrom` rules).
- **`ShortEnv`** — wrapper: `ShortEnvFrom(path, os.Environ())`.
- **`env`** — injectable `[]string` of `KEY=value` pairs for `ShortEnvFrom`.
  Process env is never mutated in this suite.
- **`alias`** — eligible env entry whose value is a single absolute path and
  whose name is a shell identifier, after skip rules (PATH-like lists, PWD,
  secrets, HOME name, home-valued vars, etc.).
- **`home` / `TildeHome`** — fallback when no alias applies; home is always
  `~`, never `$HOME`.

### Behaviors

- **Empty path** — `""` → `""` unchanged.
- **Normalize** — `filepath.Abs(path)`; on Abs error → input unchanged.
- **Build aliases** from `env`: value is a single absolute path (no `:` /
  newlines); name matches `[A-Za-z_][A-Za-z0-9_]*`; drop skip-list names and
  secret-ish names (`KEY`/`TOKEN`/`SECRET`/`PASSWORD`/`AUTH` substring).
- **Skip HOME name** — never emit `$HOME`; home always via `TildeHome`.
- **Home-valued var** — if a var’s value equals user home, do not emit `$THAT`;
  use `~`.
- **One replacement** — longest matching absolute prefix at a path-segment
  boundary → `$NAME` + remainder. Tie (same path value): shorter name, then
  alphabetical.
- **Segment boundary** — `X=/foo/ba` must not match `/foo/bar`.
- **No match / empty env** — `TildeHome(path)` (or abs outside home).
- **Not cwd-relative** — never `"."` / `"child/..."` even when under process cwd.
- **`ShortEnv`** — same rules with live `os.Environ()`; tests compare to
  `ShortEnvFrom(path, os.Environ())` only (no host-specific `$VAR` asserts).

## Decision Tree

```
short-env
├── from-replace                 [ShortEnvFrom: eligible alias replaces prefix]
│   ├── longest-prefix           nested X + AI → $AI/...
│   ├── exact-match              path == alias value → $X
│   ├── child-of-prefix          path under X → $X/rest
│   ├── tie-shorter-name         X vs PROJECT_X same value → $X
│   └── tie-alphabetical         AA vs BB same length+value → $AA
├── from-no-replace              [ShortEnvFrom: no $VAR — TildeHome / abs]
│   ├── segment-boundary         X=/foo/ba, path=/foo/bar → no $X
│   ├── empty-env                env=[] under home → ~/...
│   ├── nil-env                  env=nil under home → ~/...
│   ├── under-home-no-alias      under home, unrelated env → ~/...
│   ├── outside-home-no-alias    temp outside home → absolute
│   ├── empty-path               "" → ""
│   └── not-cwd-relative         under cwd+home → ~/..., not Short rel
├── from-skip                    [ShortEnvFrom: env present but ineligible]
│   ├── skip-path                PATH multi-dir list ignored
│   ├── skip-pwd                 PWD not used as $PWD
│   ├── skip-secret              FOO_KEY ignored
│   ├── skip-home-name           HOME never $HOME
│   └── home-valued-var          value==home → ~ not $MYHOME
└── current                      [ShortEnv wrapper]
    └── matches-from-os-environ  ShortEnv == ShortEnvFrom(..., os.Environ())
```

### Parameter significance (high → low)

1. **Op** — `from` (`ShortEnvFrom` + injectable env) vs `current` (`ShortEnv`).
2. **Outcome class** — env replace vs no-replace fallback vs eligibility skip.
3. **Match / skip rule** — longest prefix, segment boundary, ties, skip lists.
4. **Path shape** — empty, under home, outside home, under cwd.

## Test Index

| Leaf | Op | Description |
|------|-----|-------------|
| `from-replace/longest-prefix` | from | Longer nested alias wins over shorter parent |
| `from-replace/exact-match` | from | Path equal to alias value → `$X` |
| `from-replace/child-of-prefix` | from | Child of alias → `$X` + remainder |
| `from-replace/tie-shorter-name` | from | Same value, different lengths → shorter name |
| `from-replace/tie-alphabetical` | from | Same value and length → alphabetical name |
| `from-no-replace/segment-boundary` | from | Non-segment prefix does not match |
| `from-no-replace/empty-env` | from | Empty env → `TildeHome` |
| `from-no-replace/nil-env` | from | Nil env → `TildeHome` (no os.Environ magic) |
| `from-no-replace/under-home-no-alias` | from | Under home, no usable alias → `~/...` |
| `from-no-replace/outside-home-no-alias` | from | Outside home, no alias → absolute |
| `from-no-replace/empty-path` | from | Empty path unchanged |
| `from-no-replace/not-cwd-relative` | from | Never cwd-relative form |
| `from-skip/skip-path` | from | `PATH` list values ignored |
| `from-skip/skip-pwd` | from | `PWD` not emitted as `$PWD` |
| `from-skip/skip-secret` | from | Names with `KEY` (secret-ish) ignored |
| `from-skip/skip-home-name` | from | `HOME` never becomes `$HOME` |
| `from-skip/home-valued-var` | from | Var value equals home → `~` not `$VAR` |
| `current/matches-from-os-environ` | current | Wrapper equals `ShortEnvFrom` + `os.Environ()` |

## How to Run

From the **my** worktree root:

```sh
doctest vet ./external/dot-pkgs-master-2026-08-16/go-pkgs/pathfmt/tests/short-env/
doctest test ./external/dot-pkgs-master-2026-08-16/go-pkgs/pathfmt/tests/short-env/
```

From the go-pkgs module root:

```sh
doctest vet ./pathfmt/tests/short-env/
doctest test ./pathfmt/tests/short-env/
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

type Request struct {
	Op   string   // "from" (default, ShortEnvFrom) | "current" (ShortEnv)
	Path string
	Env  []string // KEY=value; only meaningful for Op=from
}

type Response struct {
	Display string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	var display string
	switch req.Op {
	case "current":
		display = pathfmt.ShortEnv(req.Path)
	default:
		display = pathfmt.ShortEnvFrom(req.Path, req.Env)
	}
	return &Response{Display: display}, nil
}
```
