# Scenario

**Feature**: shared ANSI enablement — Mode, Resolve, Style, WriterIsTTY

```
# flags pick Mode; Resolve decides; Style wraps or passes through
ModeFromFlags(color, noColor) -> Mode
Resolve(Mode, TTY, noColorEnv) -> Enabled
Style{Enabled} -> Green/Red/Yellow/Blue/Gray/Bold/Dim/Strike/Paint
WriterIsTTY(writer) -> TTY-ness for Auto
```

## Preconditions

- Package path: `github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color` (greenfield; compile-RED until implemented).
- `Run` calls `Resolve`, `ModeFromFlags`, `Style` methods, or `WriterIsTTY` from that package. No stubs.
- `noColorEnv` is injected via `Resolve(mode, tty, noColorEnv)`. Never `os.Setenv` / `t.Setenv` / `Unsetenv` / `Chdir`. Do not reassign `os.Stdout` / `os.Stderr`.
- Leaves are L2 in-process and run under `t.Parallel()`. No subprocess. No L1 `*_test.go`. No `label: e2e`.
- `Enabled(mode)` / `EnabledFor(mode, w)` are product wrappers around Resolve + `os.Getenv("NO_COLOR")` and are not doctest targets.

## Steps

1. Branch `Setup` sets `req.Op` (`resolve` | `flags` | `style` | `tty`) and the inputs for that surface.
2. Root `Run` calls the matching public API and records `Response`.
3. Leaf `Assert` checks the bool, Mode, exact SGR string, TTY-ness, or exact error text.

## Context

- **Resolve**: Always → true; Never → false; Auto: non-empty `noColorEnv` → false, else `stdoutIsTTY`.
- **ModeFromFlags**: `(false,false)=Auto`, `(true,false)=Always`, `(false,true)=Never`, `(true,true)` error exactly `--color and --no-color cannot be specified together`.
- **Style**: `!Enabled` or `text==""` → return `text` unchanged. Else wrap:
  - Green `\x1b[32m` + text + `\x1b[0m`
  - Red `\x1b[31m` + text + `\x1b[0m`
  - Yellow `\x1b[33m` + text + `\x1b[0m`
  - Gray `\x1b[90m` + text + `\x1b[0m`
  - Blue `\x1b[34m` + text + `\x1b[0m`
  - Bold `\x1b[1m` + text + `\x1b[0m`
  - Dim `\x1b[2m` + text + `\x1b[0m`
  - Strike `\x1b[9m` + text + `\x1b[0m`
  - `Paint(text, Gray, Strike)` → `\x1b[90m\x1b[9m` + text + `\x1b[0m`
- **WriterIsTTY**: `bytes.Buffer` and `os.Pipe` writer → false. Do not require a real TTY.
- Out of scope: `FORCE_COLOR` / `CLICOLOR`, 256-color / truecolor, magenta/cyan, CSI cursor/erase (`terminal/cursor`), `StripANSI`, child-process env, domain status tokens.
