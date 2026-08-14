# Scenario

**Feature**: user asked for usage; no rewrite

```
RunCLI --help|-h -> usage on stdout -> exit 0
```

## Preconditions

- No repository is required.
- Help is a short in-process CLI path.

## Steps

1. Leaf sets `req.Args` to a help flag.
2. `RunCLI` prints the locked usage text and returns nil.
