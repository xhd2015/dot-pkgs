# Scenario

**Feature**: a complete CSI 6n query in one buffer yields CPR from injected cursor

```
# complete query
buf contains ESC[6n (optionally with surrounding noise / multiples)
  -> consumeCSI6nQueries
  -> replies = ESC[<row>;<col>R … ; rest empty
```

## Preconditions

- `req.Phase = "consume"`.
- Input is a single complete buffer (`req.Data`); no chunking.

## Steps

1. Set phase to consume.
2. Leaves set `Data` and `Row`/`Col`.

## Context

Validates pure detection + reply formatting without write-func indirection.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "consume"
	return nil
}
```
