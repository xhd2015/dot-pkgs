# Scenario

**Feature**: well-formed `\xHH` input decodes to the matching raw bytes

```
# two hex digits per step, ParseInt base 16
caller \xHH\xHH… -> Decode -> one byte per step
```

## Preconditions

- Input length is a multiple of 4 and every step is `\x` plus two hex digits (`0-9a-fA-F`).
- `err` must be nil. `resp.Decoded` is the raw bytes (not UTF-8 runes).

## Steps

1. Leaf sets a well-formed `req.Hex`.
2. Assert compares `resp.Decoded` with `bytes.Equal`.
