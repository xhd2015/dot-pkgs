# Scenario

**Feature**: skip empty name, root `/`, non-`n` lines, and non-absolute names

```
junk -Fn blob -> ParseLsofFn -> only real absolute paths (not n/ or relative)
```

## Steps

1. Set `req.LsofOutput` with junk and one keep path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// p/f lines, empty n, n/, relative n, then one keep
	req.LsofOutput = []byte("p99\nfcwd\nn\nn/\nnrelative\nn./foo\nn/tmp/keep-me\nf1\nn/\n")
	return nil
}
```
