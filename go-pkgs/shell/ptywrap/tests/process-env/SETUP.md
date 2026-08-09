# Scenario

**Feature**: PTY spawn builds child `cmd.Env` via pure merge of base + set/unset, then optional spawn TERM policy

```
# pure merge (L2 product helper)
base []string + Unset keys + Set KEY=value
  -> MergeProcessEnv
  -> result environ (last-wins on set; unset removes keys)

# spawn path TERM
merged environ -> EnsureSpawnTERM
  -> if TERM missing|empty|dumb then TERM=xterm-256color
  -> else leave TERM unchanged

# production wire-up (not spawned in this tree)
SpawnOptions.Env / Unset + os.Environ()
  -> EnsureSpawnTERM(MergeProcessEnv(...))
  -> then ExtraPaths / PS1 appends (unchanged, out of scope)
```

## Preconditions

1. Package `github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap` is importable.
2. Implementer provides pure helpers (RED until present):
   - `MergeProcessEnv(base, set, unset []string) []string`
   - `EnsureSpawnTERM(env []string) []string`
3. Tests pass explicit `base` slices — **no** `t.Setenv`, `os.Setenv`, `t.Chdir`,
   or process-global environ mutation (parallel-safe).
4. ExtraPaths / PS1 remain separate post-merge appends in `startPTY`; this tree
   does not re-test PATH concatenation.

## Steps

1. Grouping nodes set `req.ApplySpawnTERM` (merge vs spawn-term).
2. Leaves set `req.Base`, `req.Set`, and `req.Unset`.
3. Root `Run` calls `MergeProcessEnv` and optionally `EnsureSpawnTERM`.
4. Assert inspects `resp.Env` via key lookup helpers (last-wins parse).

## Context

- **Env format**: standard `KEY=value` strings; keys are case-sensitive (Unix).
- **Unset before set**: locked product order so reintroduction after unset works.
- **TERM default value**: exactly `xterm-256color` (matches today's hardcoded
  append in `spawn.go`).
- **Good TERM**: any value that is present, non-empty, and not `dumb` — including
  `xterm-256color` itself and alternatives such as `screen-256color`.
- **S9 ExtraPaths**: still applied after env construction in spawn; not covered
  as a process spawn leaf here.

```go
import (
	"strings"
)

// envGet returns the last value for key in a KEY=value environ slice.
func envGet(env []string, key string) (string, bool) {
	prefix := key + "="
	found := false
	val := ""
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			found = true
			val = e[len(prefix):]
		} else if e == key {
			// bare key without '='; treat as empty value
			found = true
			val = ""
		}
	}
	return val, found
}

// envHas reports whether key appears in env (any KEY= or bare KEY).
func envHas(env []string, key string) bool {
	_, ok := envGet(env, key)
	return ok
}

// envAsMap last-wins parse of KEY=value entries.
func envAsMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, e := range env {
		if i := strings.IndexByte(e, '='); i >= 0 {
			m[e[:i]] = e[i+1:]
		} else if e != "" {
			m[e] = ""
		}
	}
	return m
}

// envMapsEqual compares two environ slices as last-wins key maps.
func envMapsEqual(a, b []string) bool {
	ma, mb := envAsMap(a), envAsMap(b)
	if len(ma) != len(mb) {
		return false
	}
	for k, v := range ma {
		if mb[k] != v {
			return false
		}
	}
	return true
}
```
