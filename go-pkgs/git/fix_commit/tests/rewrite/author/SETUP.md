# Scenario

**Feature**: author-only rewrite (no `-m`, no `--strip-co-author`)

```
RunCLI --name/--email -> new commit, same message, requested author fields changed
```

## Preconditions

- Message is still `fix typo`.
- Unrequested author field and both dates stay.
- Committer identity stays.

## Steps

1. Leaf appends `--name` and/or `--email`.
2. Assert new SHA, preserved message, and moved `master`.
