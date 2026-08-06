# shell/applescript — AppleScript string helpers and write-text limits

## Version

0.0.2

# DSN (Domain Specific Notion)

Package under test:

`github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript`

## Participants

- **Caller** — builds AppleScript or iTerm `write text` payloads (e.g. FollowUp shell lines).
- **EscapeString** — escapes `\` and `"` for double-quoted AppleScript literals (same rules as legacy iterm2 helpers).
- **CheckWriteText** — pure risk check for text about to be embedded in `write text "…"`.
- **DocumentWriteTextLimitation** — multi-line prose documenting empirical limits.
- **Constants** — `WriteTextSafeMaxBytes` (900), `WriteTextSoftMaxBytes` (1024).

## Write-text limit (empirical)

iTerm ForceNew delivers FollowUp commands via AppleScript:

```applescript
write text "<command>"
```

Lab (2026-08-06): **follow command byte length** ≲ 950 reliable; ≳ 1050 often EMPTY/MISMATCH/UTF-8 corruption. Multi-KB Chinese is fine when write text stays short (`bash script.sh` + body on disk).

Re-measure: `tests/scripts/measure-write-text-limit` (live iTerm; not default CI).

## Decision Tree

```text
applescript
|-- escape
|   `-- backslash-and-quote      # EscapeString \ and "
|-- check
|   |-- short-chinese-ok         # UTF-8 short → OK
|   |-- under-safe-max           # exactly SafeMax → OK
|   |-- near-limit               # SafeMax+1 .. SoftMax → NearLimit
|   |-- over-soft-max            # SoftMax+1 → SoftExceeded
|   `-- open-inject-shaped-risk  # long inject-shaped → SoftExceeded
|-- document
|   `-- limitation-mentions      # DocumentWriteTextLimitation non-empty keywords
`-- live                         # label: e2e — darwin + iTerm
    |-- short-follow-large-body  # short bash script + multi-KB 中文 → PASS
    `-- long-follow-over-soft    # long printf FollowUp over SoftMax → not exact match
```

## How to Run

```sh
cd go-pkgs   # this module root

doctest vet ./shell/applescript/tests
doctest test ./shell/applescript/tests
# live e2e (opens iTerm windows):
doctest test --label e2e ./shell/applescript/tests

# re-measure limits (manual):
go run ./shell/applescript/tests/scripts/measure-write-text-limit
```

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

// Request configures one pure or live surface.
type Request struct {
	// Op: escape | check | document | live-short | live-long
	Op string

	// EscapeInput for Op=escape
	EscapeInput string

	// CheckInput for Op=check (if empty, built from CheckPad / CheckChinese)
	CheckInput string
	// CheckPad: if CheckInput empty, build pad of this many ASCII 'X' (or 字 if Chinese)
	CheckPad     int
	CheckChinese bool
	// CheckExactLen: if >0, build string of exactly this many 'A' bytes
	CheckExactLen int

	// Live: unused fields reserved
}

// Response captures pure API outputs and live delivery results.
type Response struct {
	Escaped string

	CheckOK           bool
	CheckSoftExceeded bool
	CheckNearLimit    bool
	CheckByteLen      int
	CheckReasons      []string

	Doc string

	// Live
	LiveMatch  bool
	LiveGotLen int
	LiveWantLen int
	LiveSkipped bool
	LiveSkipReason string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	_ = d
	resp := &Response{}
	switch req.Op {
	case "escape":
		resp.Escaped = applescript.EscapeString(req.EscapeInput)
		return resp, nil
	case "check":
		in := req.CheckInput
		if in == "" {
			if req.CheckExactLen > 0 {
				in = strings.Repeat("A", req.CheckExactLen)
			} else if req.CheckChinese {
				in = strings.Repeat("字", req.CheckPad)
			} else if req.CheckPad > 0 {
				in = strings.Repeat("X", req.CheckPad)
			}
		}
		c := applescript.CheckWriteText(in)
		resp.CheckOK = c.OK
		resp.CheckSoftExceeded = c.SoftExceeded
		resp.CheckNearLimit = c.NearLimit
		resp.CheckByteLen = c.ByteLen
		resp.CheckReasons = append([]string(nil), c.Reasons...)
		return resp, nil
	case "document":
		resp.Doc = applescript.DocumentWriteTextLimitation()
		return resp, nil
	case "live-short", "live-long":
		if runtime.GOOS != "darwin" {
			resp.LiveSkipped = true
			resp.LiveSkipReason = "not darwin"
			t.Skip(resp.LiveSkipReason)
		}
		if !iterm2.IsInstalled() {
			resp.LiveSkipped = true
			resp.LiveSkipReason = "iTerm2 not installed"
			t.Skip(resp.LiveSkipReason)
		}
		return runLive(t, req, resp)
	default:
		return nil, fmt.Errorf("unknown Op %q", req.Op)
	}
}

func runLive(t *testing.T, req *Request, resp *Response) (*Response, error) {
	t.Helper()
	cwd, err := os.UserHomeDir()
	if err != nil {
		return resp, err
	}
	tmp, err := os.MkdirTemp("", "applescript-live-*")
	if err != nil {
		return resp, err
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })

	gotPath := filepath.Join(tmp, "out.got")
	var payload string
	var follow string

	if req.Op == "live-short" {
		// Multi-KB Chinese body; short write text.
		payload = "HDR use <<'EOF'\n" + strings.Repeat("字", 800) + "\n[image] /tmp/MID__seq_1.png\n[image] /tmp/MID__seq_2.png\n我看下\n"
		script := filepath.Join(tmp, "run.sh")
		body := "#!/bin/bash\nprintf %s " + shellQuote(payload) + " > " + shellQuote(gotPath) + "\n"
		if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
			return resp, err
		}
		follow = "bash " + shellQuote(script)
		if !applescript.CheckWriteText(follow).OK {
			t.Fatalf("control follow must be under SafeMax: len=%d", len(follow))
		}
	} else {
		// Long FollowUp over SoftMax (same shape as production open inject delivery).
		payload = "HDR use <<'EOF'\n" + strings.Repeat("字", 500) + "\n[image] /tmp/MID__seq_1.png\n[image] /tmp/MID__seq_2.png\n尾\n"
		follow = fmt.Sprintf("printf %%s %s > %s", shellQuote(payload), shellQuote(gotPath))
		if len(follow) <= applescript.WriteTextSoftMaxBytes {
			// pad payload until follow exceeds SoftMax
			for len(follow) <= applescript.WriteTextSoftMaxBytes {
				payload += strings.Repeat("字", 50) + "\n"
				follow = fmt.Sprintf("printf %%s %s > %s", shellQuote(payload), shellQuote(gotPath))
			}
		}
		if !applescript.CheckWriteText(follow).SoftExceeded {
			t.Fatalf("live-long follow should SoftExceed: len=%d", len(follow))
		}
	}

	resp.LiveWantLen = len(payload)
	if err := iterm2.OpenConfig(cwd, &iterm2.Config{
		Mode:             iterm2.ModeForceNew,
		FollowUpCommands: []string{follow},
	}); err != nil {
		return resp, err
	}

	deadline := time.Now().Add(8 * time.Second)
	var got []byte
	for time.Now().Before(deadline) {
		time.Sleep(150 * time.Millisecond)
		b, err := os.ReadFile(gotPath)
		if err == nil && len(b) > 0 {
			got = b
			time.Sleep(200 * time.Millisecond)
			if b2, err2 := os.ReadFile(gotPath); err2 == nil && len(b2) >= len(b) {
				got = b2
			}
			break
		}
	}
	resp.LiveGotLen = len(got)
	resp.LiveMatch = string(got) == payload
	_ = utf8.ValidString(payload)
	return resp, nil
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
```
