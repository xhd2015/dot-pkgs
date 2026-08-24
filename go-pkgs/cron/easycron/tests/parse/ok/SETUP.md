# Scenario

**Feature**: well-formed expressions parse successfully

```
Parse(valid) -> Expr, err=nil
```

## Preconditions

- Parent sets `req.Op` to `"parse"`.
- Leaves set `req.Expr` to a valid expression.

## Steps

1. Leaf sets `req.Expr`.
2. Assert checks Interval / Align / Until / Quiet.
