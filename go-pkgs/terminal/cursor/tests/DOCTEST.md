# terminal/cursor — CR, erase line, cursor up

L2 in-process doctests for `github.com/xhd2015/dot-pkgs/go-pkgs/terminal/cursor`.
String builders only; callers decide TTY / Interactive. Not gated by `NO_COLOR`.

## Version

0.0.2

# DSN (Domain Specific Notion)

CLI in-place progress needs a few CSI / CR sequences. This package builds those
strings. It does not detect TTY and does not emit color.

**Participants**

- **CR** — `"\r"`.
- **ClearLine** — `"\x1b[2K"` (erase entire line).
- **Up(n)** — `"\x1b[{n}A"` when n ≥ 1; `""` when n < 1.
- **Rewrite(text)** — `CR + ClearLine + text`.
- **Clear()** — `CR + ClearLine`.

**Behaviors**

- Pure strings; no `io.Writer`, no env.
- `Up(0)` and `Up(-1)` are empty (do not emit `ESC[0A`).

## Decision Tree

```
terminal/cursor/tests/
├── up-two/          Up(2) → \x1b[2A
├── up-zero/         Up(0) → ""
├── rewrite/         Rewrite("hi") → \r\x1b[2Khi
└── clear/           Clear() → \r\x1b[2K
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `up-two` | `Up(2)` is `\x1b[2A` |
| 2 | `up-zero` | `Up(0)` is empty |
| 3 | `rewrite` | `Rewrite("hi")` is CR + ClearLine + `hi` |
| 4 | `clear` | `Clear()` is CR + ClearLine |

## How to Run

From the `go-pkgs` module root:

```sh
doctest vet ./terminal/cursor/tests
doctest test ./terminal/cursor/tests
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/cursor"
)

type Request struct {
	Op   string // up | rewrite | clear
	N    int
	Text string
}

type Response struct {
	Out string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	switch req.Op {
	case "up":
		return &Response{Out: cursor.Up(req.N)}, nil
	case "rewrite":
		return &Response{Out: cursor.Rewrite(req.Text)}, nil
	case "clear":
		return &Response{Out: cursor.Clear()}, nil
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}
```
