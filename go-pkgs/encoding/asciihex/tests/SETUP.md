# Scenario

**Feature**: shared `\xHH` encode/decode so spl and kool share one core

```
# Encode each byte; Decode walks \xHH steps
caller bytes -> Encode -> \xHH\xHH… (lowercase, no newline)
caller \xHH string -> Decode -> bytes or kool-shaped error
Encode then Decode -> original bytes
```

## Preconditions

- Package path: `github.com/xhd2015/dot-pkgs/go-pkgs/encoding/asciihex` (greenfield; compile-RED until implemented).
- `Run` calls `asciihex.Encode` and/or `asciihex.Decode`. No stubs. No `kool` spawn.
- Planned API:
  - `func Encode(data []byte) string`
  - `func Decode(s string) ([]byte, error)`
- Error strings match kool `tools/encoding` `decodeAsciiHex`:
  - `invalid hex escape sequence`
  - `malformed hex escape sequence at position %d`
  - `invalid hex value %s: %v` (`strconv.ParseInt` base 16, bitSize 32)
- Leaves are L2 in-process and run under `t.Parallel()`. No process env, cwd, or stdio hijack. No L1 `*_test.go`. No `label: e2e`.

## Steps

1. Branch `Setup` sets `req.Op` (`encode` | `decode` | `roundtrip`).
2. Leaf `Setup` sets `req.Data` and/or `req.Hex`.
3. Root `Run` calls the matching public API and records `Response`.
4. Leaf `Assert` checks the exact string, raw bytes, or exact error text.

## Context

- Encode format is `fmt.Sprintf("\\x%02x", b)` per byte, concatenated. Library must **not** append `\n` (kool CLI prints a newline after the core).
- Decode is a byte decoder: `\xff` → `{0xff}`, not UTF-8 of U+00FF (`c3 bf`). kool's `WriteRune` string helper is not the library contract.
- Out of scope this phase: kool CLI wrap, `spl logs tokens update`, `domain_tokens.go`, `ProbeOpenAPI`, live OpenAPI, base64/url/plain-hex algorithms.
