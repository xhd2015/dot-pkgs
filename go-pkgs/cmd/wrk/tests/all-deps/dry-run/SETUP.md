# Scenario

**Feature**: wrk --all-deps --dry-run plans every matched dependency but writes nothing

```
# consumer requires deps; scan-root has matching repos -> wrk --all-deps --dry-run prints would: lines, touches nothing
consumer (go.mod + git) + scan-root (mydep1, mydep2) -> wrk --all-deps --dry-run --scan-root <root> -> stdout "would: ..." lines, no external/, no replace, no .gitignore
# bare --dry-run without --all-deps -> error before any planning
wrk --dry-run -> error (--dry-run is only valid with --all-deps)
```

## Preconditions

- Git and Go must be available.
- Consumer cwd must be inside a git work tree with a `go.mod`.
- Dep repos under the scan root must be `RepoTypeMain` on branch `main` with a committed `go.mod`.
- `WRK_DATE=2026-06-30` (existing `wrkDate`) flows via root `Run`; expected external names use `2026-06-30`.

## Steps

- Each leaf builds an isolated consumer git repo plus named dep repos under a temp scan root, reusing the inherited `initAllDepsConsumer` / `initAllDepsRepo` helpers.
- `req.Args` carries `--dry-run` as a plain arg (no `Request` field); `--all-deps` + `--dry-run` + `--scan-root <root>` for the planning leaves, bare `--dry-run` for the error leaf.
- The `allDeps*` helpers (`allDepsReadGoMod`, `allDepsHasReplaceForModule`, `allDepsGitignoreContainsExternal`, `allDepsExternalRelPath`, `allDepsExternalAbsPath`, ...) are inherited from `all-deps/SETUP.md` and are NOT redefined here.

## Context

- Dry-run stdout uses relative `./external/<name>` paths, so no `EvalSymlinks` is needed for stdout assertions; the inherited `allDepsExternalAbsPath` already normalizes via `EvalSymlinks` for any abs-path check.
- The core dry-run guarantee is asserted strongly in the planning leaves: `external/` absent, `go.mod` `Replace` list empty (or no replace for the deps), `.gitignore` has no `/external` line.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	// The allDeps* helpers are inherited from all-deps/SETUP.md; keep the
	// dryRun-only symbols referenced so the inlined per-leaf test func compiles.
	dryRunEnsureHelpersUsed()
	return nil
}

// dryRunEnsureHelpersUsed keeps the inherited allDeps* helpers and any dry-run
// assertions referenced even when a given leaf does not call every one (avoids
// unused-symbol compile errors in the inlined per-leaf test func).
func dryRunEnsureHelpersUsed() {
	_ = allDepsGoModJSON{}
	_ = allDepsReadGoMod
	_ = allDepsHasReplaceForModule
	_ = allDepsReplacePathForModule
	_ = allDepsGitignoreContainsExternal
	_ = allDepsCountGitignoreExternalLines
	_ = allDepsExternalRelPath
	_ = allDepsExternalAbsPath
	_ = allDepsRunGo
	_ = initAllDepsRepo
	_ = initAllDepsConsumer
	_ = allDepsEnsureHelpersUsed
}
```
