# Scenario

**Feature**: `--dry-run` prints the plan and skips mutations

```
discovery + ls-remote -> [dry-run] lines on stdout -> no commit-tree / ref / remote writes
```

## Preconditions

- Same pipeline as live: validation, strip parse, ref listing, `ls-remote`.
- No new SHA (commit-tree skipped). First line is
  `[dry-run] would rewrite <oldsha>`.
- Every planned line is prefixed with `[dry-run]` on stdout.

## Steps

1. Leaf appends `-m` and `--dry-run` (and any tag fixture).
2. Assert plan text and that refs/remotes still point at the old SHA.
