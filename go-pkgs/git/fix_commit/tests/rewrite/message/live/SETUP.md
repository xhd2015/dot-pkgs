# Scenario

**Feature**: live `-m` rewrite (side effects run)

```
RunCLI -m … -> commit-tree + update-ref [+ tag/push] -> success report
```

## Preconditions

- `--dry-run` is off. `commit-tree` produces a new SHA printed as
  `rewrote <old> -> <new>`.
- Exact-tip local branches move. Descendants are not rewritten.

## Steps

1. Leaf appends `-m` and any extra flags/topology.
2. Assert stdout report plus git object/ref side effects.
