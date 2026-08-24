# Scenario

**Feature**: human-friendly easy-cron Parse / Next / Active

```
# pipeline
caller expr string -> Parse -> Expr
caller Expr + anchor/from -> Next -> fire time | expired
caller Expr + at -> Active -> bool
```

## Preconditions

- Package path: `github.com/xhd2015/dot-pkgs/go-pkgs/cron/easycron`.
- `Run` calls public `Parse`, `Next`, and `Active` only. No stubs. No kck spawn.
- Leaves are L2 in-process and run under `t.Parallel()`. No process env, cwd, or stdio hijack. No `label: e2e`.
- All times in leaves are RFC3339 interpreted in fixed UTC.

## Steps

1. Branch `Setup` sets `req.Op` (`parse` | `next` | `active`).
2. Leaf `Setup` sets `req.Expr` and any times.
3. Root `Run` calls the matching API and records `Response`.
4. Leaf `Assert` checks fields or error text.

## Context

- Relative schedules fire at `anchor + k*interval` (k≥0); inclusive `from`.
- Aligned schedules use per-day midnight grids.
- Until deadline is exclusive on anchor's local calendar day.
- Quiet overnight: Active is false for TOD in `[Start, 24h) ∪ [0, End)`.
