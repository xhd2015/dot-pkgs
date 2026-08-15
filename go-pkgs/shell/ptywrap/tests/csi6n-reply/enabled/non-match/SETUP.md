# Scenario

**Feature**: non-6n escape sequences do not produce a CPR reply

```
# non-match
ESC[H | ESC[5n | ESC[?6n | plain text
  -> consumeCSI6nQueries
  -> replies empty; rest empty (complete non-query)
```

## Preconditions

- `req.Phase = "consume"`.
- Sequences are complete (not incomplete fragments).

## Steps

1. Set phase to consume.
2. Leaves set non-matching `Data`.

## Context

Seals: only plain `ESC[6n` is a cursor query. DEC private `?` and DSR status
`5n` must not be mistaken for CPR requests.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "consume"
	req.Row = 2
	req.Col = 2
	return nil
}
```
