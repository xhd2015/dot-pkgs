# Scenario

**Feature**: invalid expressions return errors

```
Parse(invalid) -> error
```

## Preconditions

- Parent sets `req.Op` to `"parse"`.
- Leaves set `req.Expr` to an invalid expression.

## Steps

1. Leaf sets `req.Expr`.
2. Assert checks `err != nil` (message may be checked loosely).
