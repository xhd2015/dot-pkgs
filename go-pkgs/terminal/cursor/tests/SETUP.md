# Scenario

**Feature**: cursor string builders (CR, erase line, up)

```
Up(n) / Rewrite(text) / Clear()
```

## Preconditions

- Package: `github.com/xhd2015/dot-pkgs/go-pkgs/terminal/cursor`.
- No env, no TTY, no `Setenv`. L2 in-process, `t.Parallel()`.

## Steps

1. Leaf `Setup` sets `req.Op` and inputs.
2. `Run` calls the matching function.

## Context

- `Up(n)` empty when n < 1.
- `Rewrite` / `Clear` always include CR + `\x1b[2K`.
