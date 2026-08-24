# cron/easycron — human interval expressions

## Version

0.0.1

L2 in-process doctests for `github.com/xhd2015/dot-pkgs/go-pkgs/cron/easycron`.
Pure library: Parse / Next / Active / Deadline. No ticker, no kck wiring, no L3.

## DSN (Domain Specific Notion)

### Participants

- **Parse** — string → `Expr` or error.
- **Expr** — Interval, optional Align (`-at-`), Until (`-until-`), Quiet (`-not-within-`).
- **Next** — earliest fire `>= from` that is Active and strictly before Deadline.
- **Active** — true when outside Quiet (Until ignored).
- **Deadline** — exclusive Until instant on anchor's local calendar day.
- **Anchor** — schedule start; relative grids use `anchor + k*Interval`.
- **Loc** — tests use fixed `UTC` (`time.FixedZone("UTC", 0)`).

### Behaviors

- Grammar (fixed suffix order):
  `every-<dur>[-at-<offset>][-until-<tod>][-not-within-<tod>-to-<tod>]`
- `<dur>` / `<offset>`: `Nh`, `Nm`, or `NhNm`. Offset in `[0, Interval)`.
- `<tod>`: `NhNm` with hour 0–23 and minute 0–59 (both parts required).
- Without `-at-`: first fire is `anchor` (inclusive `from`).
- With `-at-`: local wall-clock grid restarts each midnight at `midnight+align`, then `+interval`, staying before next midnight.
- `-until-`: hard stop; fires must be strictly before Deadline; starting at/after today's Until → expired.
- `-not-within-A-to-B`: recurring quiet; overnight when A > B; quiet includes Start, Active at End.

### Inverse

No inverse. Scheduling is not a codec.

## Decision Tree

```
cron/easycron/tests/
├── parse/                           Parse → Expr | error
│   ├── ok/                          well-formed
│   │   ├── every-1h/
│   │   ├── every-1h-at-4m/
│   │   ├── every-2h-at-90m/
│   │   ├── every-5m-until-19h00m/
│   │   ├── every-5m-not-within/
│   │   └── compose-all/
│   └── err/                         invalid input
│       ├── empty/
│       ├── bad-prefix/
│       ├── offset-ge-interval/
│       ├── until-bad-hour/
│       └── not-within-missing-to/
├── next/                            Next(anchor, from, loc)
│   ├── relative-includes-now/
│   ├── relative-sequence/
│   ├── aligned-snaps-forward/
│   ├── until-last-before-deadline/
│   ├── until-expired-at-start/
│   └── quiet-overnight-resume/
└── active/                          Active(at, loc)
    ├── in-quiet-evening/
    ├── at-quiet-end-active/
    └── no-quiet-always/
```

### Parameter significance (high → low)

1. **Op** — parse vs next vs active.
2. **Validity** (parse) — ok vs error class.
3. **Grid mode** (next) — relative vs aligned vs until vs quiet.
4. **Boundary** — inclusive now, exclusive until, quiet end resume.

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `parse/ok/every-1h` | Parse `every-1h` → Interval 1h, no align/until/quiet |
| 2 | `parse/ok/every-1h-at-4m` | Align 4m |
| 3 | `parse/ok/every-2h-at-90m` | Align 90m OK for 2h |
| 4 | `parse/ok/every-5m-until-19h00m` | Until 19:00 |
| 5 | `parse/ok/every-5m-not-within` | Quiet 19:00→06:30 |
| 6 | `parse/ok/compose-all` | at + until + not-within |
| 7 | `parse/err/empty` | empty → error |
| 8 | `parse/err/bad-prefix` | `daily-1h` → error |
| 9 | `parse/err/offset-ge-interval` | `every-1h-at-60m` → error |
| 10 | `parse/err/until-bad-hour` | `until-25h00m` → error |
| 11 | `parse/err/not-within-missing-to` | missing `-to-` → error |
| 12 | `next/relative-includes-now` | `every-1h` at anchor → anchor |
| 13 | `next/relative-sequence` | next after anchor → +1h |
| 14 | `next/aligned-snaps-forward` | 10:07 → 11:04 |
| 15 | `next/until-last-before-deadline` | last fire 18:55; none at 19:00 |
| 16 | `next/until-expired-at-start` | start 20:00 until-19h → no fire |
| 17 | `next/quiet-overnight-resume` | from 19:01 → >= next 06:30 active |
| 18 | `active/in-quiet-evening` | 19:01 quiet |
| 19 | `active/at-quiet-end-active` | 06:30 active |
| 20 | `active/no-quiet-always` | `every-1h` always active |

## How to Run

From the go-pkgs module root:

```sh
doctest vet ./cron/easycron/tests
doctest test ./cron/easycron/tests
```

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/cron/easycron"
)

// Request drives one L2 scenario against the public easycron API.
// Op selects the surface: parse | next | active.
type Request struct {
	Op     string
	Expr   string
	Anchor string // RFC3339, required for next
	From   string // RFC3339, required for next
	At     string // RFC3339, required for active
}

// Response holds observed package outputs for Assert.
type Response struct {
	Expr     easycron.Expr
	Next     time.Time
	NextOK   bool
	Active   bool
	ParseErr string // set when Op=parse and Parse fails (err returned too)
}

func utc() *time.Location {
	return time.FixedZone("UTC", 0)
}

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return tm.In(utc())
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	switch req.Op {
	case "parse":
		expr, err := easycron.Parse(req.Expr)
		if err != nil {
			return &Response{ParseErr: err.Error()}, err
		}
		return &Response{Expr: expr}, nil
	case "next":
		expr, err := easycron.Parse(req.Expr)
		if err != nil {
			return nil, err
		}
		next, ok := expr.Next(mustTime(t, req.Anchor), mustTime(t, req.From), utc())
		return &Response{Expr: expr, Next: next, NextOK: ok}, nil
	case "active":
		expr, err := easycron.Parse(req.Expr)
		if err != nil {
			return nil, err
		}
		return &Response{Expr: expr, Active: expr.Active(mustTime(t, req.At), utc())}, nil
	default:
		t.Fatalf("unknown Op %q", req.Op)
		return nil, nil
	}
}
```
