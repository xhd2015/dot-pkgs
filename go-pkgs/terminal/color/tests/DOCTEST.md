# terminal/color — shared ANSI enablement + Style

L2 in-process doctests for `github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color`.
Classic TDD: the package does not exist yet; `Run` calls the real APIs so the
suite is compile-RED until the implementer lands it.

## Version

0.0.2

# DSN (Domain Specific Notion)

Shared terminal coloring so CLIs agree on when to wrap text in SGR and which
sequences to emit. Flags pick a **Mode**; **Resolve** turns Mode + TTY +
**noColorEnv** into a bool; **Style** wraps or passes through; **WriterIsTTY**
reports writer TTY-ness for Auto.

**Participants**

- **Mode** — `Auto` (default), `Always` (`--color`), `Never` (`--no-color`).
- **ModeFromFlags** — maps the two bool flags to Mode, or an exclusive-or error.
- **Resolve** — `(Mode, stdoutIsTTY, noColorEnv) -> bool`; injectable (no process env).
- **noColorEnv** — injected `NO_COLOR` value; meaningful only in Auto.
- **Style** — `{Enabled bool}` plus `Green` / `Red` / `Yellow` / `Blue` / `Gray` / `Bold` / `Dim` / `Strike` / `Paint`.
- **WriterIsTTY** — `io.Writer` → whether that writer is a TTY (no real TTY required).

**Behaviors**

- Always → color on; Never → color off; flags win over `NO_COLOR`.
- Auto: any non-empty `noColorEnv` disables; empty `noColorEnv` follows TTY.
- `ModeFromFlags(true, true)` errors: `--color and --no-color cannot be specified together`.
- Style: `!Enabled` or empty text → return text unchanged; else SGR + `\x1b[0m`.
- `WriterIsTTY` is false for `bytes.Buffer` and `os.Pipe` writers.

`Enabled` / `EnabledFor` wrap Resolve + env/TTY and are not leaves.

## Decision Tree

```
terminal/color/tests/
├── resolve/                         Mode × TTY × noColorEnv
│   ├── always-non-tty/              Always + !tty → true
│   ├── always-with-no-color/        Always + tty + "1" → true (flags win)
│   ├── never-tty/                   Never + tty → false
│   ├── auto-tty/                    Auto + tty → true
│   ├── auto-pipe/                   Auto + !tty → false
│   ├── auto-no-color/               Auto + tty + "1" → false
│   └── auto-empty-no-color/         Auto + tty + "" → true
├── flags/                           ModeFromFlags 2×2
│   ├── neither/                     false,false → Auto
│   ├── color-only/                  true,false → Always
│   ├── no-color-only/               false,true → Never
│   └── conflict/                    true,true → exact error
├── style/                           wrap vs passthrough
│   ├── disabled-green/              Enabled=false → plain text
│   ├── enabled-green/               exact \x1b[32m … \x1b[0m
│   ├── enabled-red/                 exact \x1b[31m … \x1b[0m
│   ├── enabled-yellow/              exact \x1b[33m … \x1b[0m
│   ├── enabled-gray/                exact \x1b[90m … \x1b[0m
│   ├── enabled-blue/                exact \x1b[34m … \x1b[0m
│   ├── enabled-bold/                exact \x1b[1m … \x1b[0m
│   ├── enabled-dim/                 exact \x1b[2m … \x1b[0m
│   ├── enabled-strike/              exact \x1b[9m … \x1b[0m
│   ├── paint-gray-strike/           Paint(Gray, Strike) one reset
│   └── empty-text/                  Enabled=true, "" → ""
└── tty/                             WriterIsTTY writer kind
    ├── buffer-not-tty/              bytes.Buffer → false
    └── pipe-not-tty/                os.Pipe writer → false
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `resolve/always-non-tty` | `Always` + non-TTY + empty env → enabled |
| 2 | `resolve/always-with-no-color` | `Always` + TTY + `"1"` → enabled (flags win) |
| 3 | `resolve/never-tty` | `Never` + TTY → disabled |
| 4 | `resolve/auto-tty` | `Auto` + TTY + empty env → enabled |
| 5 | `resolve/auto-pipe` | `Auto` + non-TTY + empty env → disabled |
| 6 | `resolve/auto-no-color` | `Auto` + TTY + `"1"` → disabled |
| 7 | `resolve/auto-empty-no-color` | `Auto` + TTY + `""` → enabled (empty does not disable) |
| 8 | `flags/neither` | no flags → `Auto`, nil error |
| 9 | `flags/color-only` | `--color` → `Always` |
| 10 | `flags/no-color-only` | `--no-color` → `Never` |
| 11 | `flags/conflict` | both flags → error `--color and --no-color cannot be specified together` |
| 12 | `style/disabled-green` | `Enabled=false` Green keeps plain text |
| 13 | `style/enabled-green` | Green wraps with `\x1b[32m` + reset |
| 14 | `style/enabled-red` | Red wraps with `\x1b[31m` + reset |
| 15 | `style/enabled-yellow` | Yellow wraps with `\x1b[33m` + reset |
| 16 | `style/enabled-gray` | Gray wraps with `\x1b[90m` + reset |
| 17 | `style/enabled-blue` | Blue wraps with `\x1b[34m` + reset |
| 18 | `style/enabled-bold` | Bold wraps with `\x1b[1m` + reset |
| 19 | `style/enabled-dim` | Dim wraps with `\x1b[2m` + reset |
| 20 | `style/enabled-strike` | Strike wraps with `\x1b[9m` + reset |
| 21 | `style/paint-gray-strike` | `Paint(Gray, Strike)` is gray+strike+one reset |
| 22 | `style/empty-text` | enabled + empty string → empty (no escape pair) |
| 23 | `tty/buffer-not-tty` | `bytes.Buffer` is not a TTY |
| 24 | `tty/pipe-not-tty` | `os.Pipe` writer is not a TTY |

## How to Run

From the `go-pkgs` module root:

```sh
doctest vet ./terminal/color/tests
doctest test ./terminal/color/tests
```

```go
import (
	"bytes"
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/color"
)

// Request drives one L2 scenario against the public color API.
// Op selects the surface: resolve | flags | style | tty.
type Request struct {
	Op string

	// resolve
	Mode       color.Mode
	TTY        bool
	NoColorEnv string

	// flags
	ColorFlag   bool
	NoColorFlag bool

	// style
	Enabled bool
	Color   string // green | red | yellow | gray | blue | bold | dim | strike | paint-gray-strike
	Text    string

	// tty
	WriterKind string // buffer | pipe
}

// Response holds observed package outputs for Assert.
type Response struct {
	Enabled bool
	Mode    color.Mode
	Out     string
	IsTTY   bool
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	resp := &Response{}

	switch req.Op {
	case "resolve":
		resp.Enabled = color.Resolve(req.Mode, req.TTY, req.NoColorEnv)
		return resp, nil

	case "flags":
		mode, err := color.ModeFromFlags(req.ColorFlag, req.NoColorFlag)
		resp.Mode = mode
		return resp, err

	case "style":
		s := color.Style{Enabled: req.Enabled}
		switch req.Color {
		case "green":
			resp.Out = s.Green(req.Text)
		case "red":
			resp.Out = s.Red(req.Text)
		case "yellow":
			resp.Out = s.Yellow(req.Text)
		case "gray":
			resp.Out = s.Gray(req.Text)
		case "blue":
			resp.Out = s.Blue(req.Text)
		case "bold":
			resp.Out = s.Bold(req.Text)
		case "dim":
			resp.Out = s.Dim(req.Text)
		case "strike":
			resp.Out = s.Strike(req.Text)
		case "paint-gray-strike":
			resp.Out = s.Paint(req.Text, color.Gray, color.Strike)
		default:
			t.Fatalf("unknown Color %q", req.Color)
		}
		return resp, nil

	case "tty":
		switch req.WriterKind {
		case "buffer":
			var buf bytes.Buffer
			resp.IsTTY = color.WriterIsTTY(&buf)
		case "pipe":
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("os.Pipe: %v", err)
			}
			t.Cleanup(func() {
				_ = r.Close()
				_ = w.Close()
			})
			resp.IsTTY = color.WriterIsTTY(w)
		default:
			t.Fatalf("unknown WriterKind %q", req.WriterKind)
		}
		return resp, nil

	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}
```
