# Lossless Markdown section editing

Classic-TDD scenarios for the in-memory `markdown.Document` API. The suite is
RED until `github.com/xhd2015/dot-pkgs/go-pkgs/markdown` implements the public
contract imported by the harness.

## Version

0.0.2

# DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies source Markdown, an ATX header selector, and optionally replacement or insertion content.
- **Document** — retains source bytes and exposes lookup and mutation operations.
- **Section** — an ATX heading plus its body through the next equal-or-higher-level ATX heading.
- **Fence** — a backtick or tilde code fence that suppresses heading recognition until a matching close.

### Behaviors

- **Caller → Document** parses source and selects a section by exact heading level and normalized, case-sensitive title.
- **Document → Section** returns, replaces, or removes the selected body without changing unrelated bytes.
- **Document → Caller** reports missing, existing, ambiguous, and invalid-selector outcomes with sentinel errors.
- **Fence → Document** makes heading-looking code literal rather than a section boundary.

## Decision Tree

```text
section-editing/
├── lookup/
│   ├── start-with-child-and-same-level
│   ├── middle-with-higher-level-boundary
│   ├── eof-empty-body
│   ├── missing-section
│   ├── backtick-fence
│   ├── tilde-fence
│   ├── unclosed-fence
│   ├── indented-closing-hashes
│   ├── case-sensitive-title
│   ├── invalid-selector
│   └── duplicate-ambiguous
├── replace/
│   ├── preserve-heading-and-mixed-newlines
│   ├── normalize-missing-content-newline
│   ├── empty-body-at-eof
│   ├── idempotent-existing-body
│   ├── missing-atomic
│   └── duplicate-atomic
├── remove/
│   ├── start
│   ├── middle
│   ├── eof
│   ├── all-duplicates
│   ├── missing-atomic
│   └── duplicate-atomic
└── insert/
    ├── before-first-heading-after-preamble
    ├── no-heading-no-final-newline
    ├── empty-source
    ├── mixed-newlines-use-first-style
    ├── empty-content
    ├── existing-atomic
    ├── duplicate-atomic
    └── invalid-selector-atomic
```

Parameter ranking (most to least significant): operation, selector resolution,
Markdown structural context, content/newline shape.

## Test Index

| Branch | Leaves | Contract |
|---|---:|---|
| `lookup` | 11 | boundaries, empty vs missing, fences, normalized headings, invalid/duplicate selectors |
| `replace` | 6 | exact preservation, newline policy including empty EOF bodies, idempotence, missing/duplicate atomicity |
| `remove` | 6 | start/middle/EOF deletion, remove-all duplicates, and atomic errors |
| `insert` | 8 | preambles, no-heading/empty sources, newline policy, empty body, error atomicity |

## How to Run

```sh
cd go-pkgs
doctest vet ./markdown/tests/section-editing
doctest test ./markdown/tests/section-editing
```

All 31 leaves are unlabeled L2 in-process tests. No filesystem, process,
environment, cwd, or shared mutable state is used.

```go
import (
	"errors"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/markdown"
)

type Request struct {
	Op      string
	Source  string
	Header  string
	Content string
}

type Response struct {
	Content      string
	Found        bool
	Removed      int
	Output       string
	SecondOutput string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	doc, err := markdown.Parse(req.Source)
	if err != nil {
		return nil, err
	}
	resp := &Response{}
	switch req.Op {
	case "lookup":
		resp.Content, resp.Found, err = doc.GetSectionContent(req.Header)
	case "replace":
		err = doc.ReplaceSectionContent(req.Header, req.Content)
	case "remove":
		err = doc.RemoveSection(req.Header)
	case "remove_all":
		resp.Removed, err = doc.RemoveAllSections(req.Header)
	case "insert":
		err = doc.InsertBeforeFirstSection(req.Header, req.Content)
	default:
		t.Fatalf("unknown operation %q", req.Op)
	}
	resp.Output = doc.String()
	resp.SecondOutput = doc.String()
	return resp, err
}

func assertExact(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("exact text mismatch\n got: %q\nwant: %q", got, want)
	}
}

func assertSentinel(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want errors.Is(_, %v)", err, want)
	}
}
```
