# Scenario

**Feature**: Decode rejects strings that are not a full `\xHH` walk

```
# prefix / leftover group / non-hex — same classes as kool decodeAsciiHex
caller bad Hex -> Decode -> exact error
```

## Preconditions

- Input fails kool's walk: missing `\x` prefix / too short, a mid-string step that is not `\xHH`, or non-hex digits.
- `err` must be non-nil. `resp.Decoded` is ignored.

## Steps

1. Leaf sets a malformed `req.Hex`.
2. Assert checks the exact error string for that class.
