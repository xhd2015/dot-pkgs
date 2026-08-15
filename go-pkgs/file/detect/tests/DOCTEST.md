# file/detect — File Type / Binary Sniff

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/file/detect`. `DetectFileType`
sniffs a path's first bytes (magic, text/UTF-8, NUL) and reports a description plus
whether the file should be treated as binary.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies a filesystem path to classify.
- **DetectFileType** — opens the path, reads up to 512 sniff bytes, returns
  `(desc string, isBinary bool, err error)`.
- **Magic catalog** — known headers (PNG, ELF, archives, …) force binary (except the
  generic `"text"` magic path).
- **UTF-8 / content sniff** — NUL ⇒ binary; valid UTF-8 text (including non-ASCII box
  drawing) ⇒ not binary; incomplete multi-byte sequence **only at the end of the sniff
  window** must still count as a valid UTF-8 prefix (truncation-safe).

### Behaviors

- Empty file → not binary.
- ASCII text → not binary (`"text"` / `"text file"`).
- UTF-8 text starting with non-ASCII (e.g. box drawing `│`) → not binary.
- Real TTY status snapshot fixture (`05-status-fields.snapshot.txt`) → not binary;
  must not return `"binary file"`.
- Sniff buffer splitting a multi-byte UTF-8 rune at offset 512 → not binary.
- NUL byte in content → binary (`"binary file"`).
- Known binary magic (PNG, …) → binary (covered by package unit tests; not re-asserted
  here unless a leaf needs it).

## Decision Tree

```
detect
└── utf8-text-not-binary/          [DetectFileType text vs binary]
    ├── real-snapshot/             production TTY snapshot fixture → not binary (RED until fix)
    ├── ascii-text/                hello world → not binary
    ├── nul-binary/                hello\x00world → binary
    └── utf8-sniff-truncation/     multi-byte rune straddling sniff boundary → not binary
```

## Test Index

| Leaf | Description |
|------|-------------|
| `utf8-text-not-binary/real-snapshot` | Internalized `05-status-fields.snapshot.txt` → `isBinary == false` |
| `utf8-text-not-binary/ascii-text` | ASCII text file → not binary |
| `utf8-text-not-binary/nul-binary` | File containing NUL → binary |
| `utf8-text-not-binary/utf8-sniff-truncation` | Valid UTF-8 with 3-byte rune split at byte 512 → not binary |

## How to Run

```sh
cd external/dot-pkgs-master-2026-07-11/go-pkgs
doctest vet ./file/detect/tests
doctest test -v ./file/detect/tests/...
go test ./file/detect/... -count=1
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/file/detect"
)

type Request struct {
	// Path is the filesystem path passed to DetectFileType.
	Path string
}

type Response struct {
	Desc     string
	IsBinary bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	desc, isBinary, err := detect.DetectFileType(req.Path)
	if err != nil {
		return nil, err
	}
	return &Response{
		Desc:     desc,
		IsBinary: isBinary,
	}, nil
}
```
