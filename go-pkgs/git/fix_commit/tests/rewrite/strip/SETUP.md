# Scenario

**Feature**: `--strip-co-author` is the message source

```
# default fixture: subject + mixed-case trailers (one with CR)
amend HEAD message -> RunCLI --strip-co-author -> leftover subject
```

## Preconditions

- A line matches if, after trim (including `\r`), it **starts with**
  `co-authored-by:` (case-insensitive, colon required).
- All matching lines are removed. A blank separator that only existed for
  those trailers is dropped. Trailing whitespace trimmed; one terminating
  newline kept.
- Default target message:

```
fix typo

Co-authored-by: Bot <bot@x>
Co-Authored-By: Other <other@x>
```

  (second trailer line ends with `\r` before the newline.)

## Steps

1. Amend HEAD to the default trailer message and refresh `req.OldSHA`.
2. Leaves that need a different message amend again.
3. Leaves append `--strip-co-author` and optional `--name` / `--dry-run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	setHEADMessage(t, req.Dir, "fix typo\n\nCo-authored-by: Bot <bot@x>\nCo-Authored-By: Other <other@x>\r\n")
	fillCommitMeta(t, req)
	return nil
}
```
