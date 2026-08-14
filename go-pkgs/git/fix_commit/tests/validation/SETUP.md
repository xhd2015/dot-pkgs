# Scenario

**Feature**: reject bad argv before planning a rewrite

```
RunCLI bad args -> Error: … -> exit 1, no git mutations
```

## Preconditions

- Flag parse, required-change-fields, and mutual exclusion run **before**
  resolving SHA. Missing SHA / missing fields / unknown flag / strip+m do
  not need a repo.
- Repo existence and revision resolution run next (`-C` / unknown SHA).

## Steps

1. Leaf sets `req.Args` to an invalid invocation.
2. `RunCLI` returns a locked `Error:` string. Harness records it on stderr.
