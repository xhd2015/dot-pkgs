# encoding/asciihex — `\xHH` encode / decode core

L2 in-process doctests for `github.com/xhd2015/dot-pkgs/go-pkgs/encoding/asciihex`.
Classic TDD: the package does not exist yet; `Run` calls the real `Encode` /
`Decode` APIs so the suite is compile-RED until the implementer lands them.

P1 extracts kool `tools/encoding` ascii_hex **core** only. Tests target the
library API, not the `kool` process. No L3. No `spl logs tokens update`.

## Version

0.0.2

# DSN (Domain Specific Notion)

Shared `\xHH` byte encoding so **spl** and **kool** emit and parse the same
ascii_hex strings without depending on each other.

**Participants**

- **Encode** — `[]byte` → string; one `\xHH` step per input byte.
- **Decode** — well-formed `\xHH\xHH…` string → `[]byte`, or error.
- **Caller** — kool wrap or spl; this tree never spawns those processes.
- **ASCII-hex string** — concatenation of 4-character `\xHH` groups; lowercase
  hex on encode; no trailing newline.
- **Bytes** — raw 0–255 values. `\xff` is the single byte `0xff`, not UTF-8 U+00FF.

**Behaviors**

- Encode `[]byte("lgu_AB")` → `\x6c\x67\x75\x5f\x41\x42` (lowercase, no `\n`).
- Encode empty / nil → `""`.
- Decode walks the whole string in steps of 4. Each step must be `\x` + two hex digits.
- Decode empty, too short, or not starting with `\x` → `invalid hex escape sequence`.
- Mid-string leftover or a step that is not `\x` → `malformed hex escape sequence at position N`.
- Non-hex digits after `\x` → `invalid hex value HH: …` (strconv `ParseInt` base 16).
- Decode is the inverse of Encode: `Decode(Encode(data))` equals `data`.

## Decision Tree

```
encoding/asciihex/tests/
├── encode/                          Encode([]byte) → string
│   ├── empty/                       []byte{} → ""
│   ├── known-lgu-ab/                lgu_AB → lowercase \xHH, no newline
│   └── binary-bytes/                0x00 0x7f 0xff → \x00\x7f\xff
├── decode/                          Decode(string) → []byte, error
│   ├── valid/                       well-formed \xHH steps
│   │   ├── a-bang/                  \x41\x21 → A!
│   │   └── uppercase-hex/           \x4A\x21 → J! (ParseInt accepts A-F)
│   └── invalid/                     kool decodeAsciiHex error classes
│       ├── empty/                   "" → invalid hex escape sequence
│       ├── missing-prefix/          "41" → invalid hex escape sequence
│       ├── truncated/               \x41\x2 → malformed at position 4
│       ├── mid-malformed/           \x41xx42 → malformed at position 4
│       └── non-hex/                 \xGG → invalid hex value GG
└── roundtrip/                       Encode then Decode
    └── mixed-bytes/                 ascii + 0x00 + 0xff invert
```

### Parameter significance (high → low)

1. **Op** — encode vs decode vs roundtrip (which public function, or both).
2. **Well-formedness** (decode) — valid walk vs prefix / group / hex-digit failure.
3. **Byte class** (encode / roundtrip) — empty, printable ASCII, high/NUL bytes.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `encode/empty` | `Encode([]byte{})` is `""` |
| 2 | `encode/known-lgu-ab` | `Encode([]byte("lgu_AB"))` is `\x6c\x67\x75\x5f\x41\x42` with no `\n` |
| 3 | `encode/binary-bytes` | `Encode([]byte{0x00,0x7f,0xff})` is `\x00\x7f\xff` |
| 4 | `decode/valid/a-bang` | `Decode(\x41\x21)` is `[]byte("A!")` |
| 5 | `decode/valid/uppercase-hex` | `Decode(\x4A\x21)` is `[]byte("J!")` |
| 6 | `decode/invalid/empty` | `Decode("")` errors `invalid hex escape sequence` |
| 7 | `decode/invalid/missing-prefix` | `Decode("41")` errors `invalid hex escape sequence` |
| 8 | `decode/invalid/truncated` | `Decode(\x41\x2)` errors `malformed hex escape sequence at position 4` |
| 9 | `decode/invalid/mid-malformed` | `Decode(\x41xx42)` errors `malformed hex escape sequence at position 4` |
| 10 | `decode/invalid/non-hex` | `Decode(\xGG)` errors `invalid hex value GG: …` |
| 11 | `roundtrip/mixed-bytes` | `Decode(Encode(data))` equals `data` including `0xff` as one raw byte |

## How to Run

From the `go-pkgs` module root:

```sh
doctest vet ./encoding/asciihex/tests
doctest test ./encoding/asciihex/tests
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/encoding/asciihex"
)

// Request drives one L2 scenario against the public asciihex API.
// Op selects the surface: encode | decode | roundtrip.
type Request struct {
	Op   string
	Data []byte // encode + roundtrip
	Hex  string // decode
}

// Response holds observed package outputs for Assert.
type Response struct {
	Encoded string
	Decoded []byte
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	switch req.Op {
	case "encode":
		return &Response{Encoded: asciihex.Encode(req.Data)}, nil
	case "decode":
		decoded, err := asciihex.Decode(req.Hex)
		return &Response{Decoded: decoded}, err
	case "roundtrip":
		encoded := asciihex.Encode(req.Data)
		decoded, err := asciihex.Decode(encoded)
		return &Response{Encoded: encoded, Decoded: decoded}, err
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}
```
