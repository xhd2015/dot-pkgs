# Scenario

**Feature**: `-m` / `--message` is the message source

```
RunCLI <sha> -m <msg> [author flags] -> new commit message is exactly <msg>
```

## Preconditions

- `--strip-co-author` is not used on this branch (mutual exclusion).
- `-m` replaces the full message (subject + body), not just the first line.

## Steps

1. Live leaves append `-m` and perform mutations.
2. Dry-run leaves append `-m --dry-run` and assert the plan only.
