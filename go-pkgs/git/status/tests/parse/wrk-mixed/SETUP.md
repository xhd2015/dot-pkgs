# Scenario

**Feature**: wrk porcelain taxonomy maps lines to five buckets with staged priority

```
unstaged M/D + ?? + staged R -> ParsePorcelainWrk
  -> Changed=1, Deleted=1, Untracked=1, Staged=1, Renamed=0
```

## Steps

1. Set `req.Op` to `"parse-wrk"`.
2. Set porcelain: unstaged modify/delete, untracked, and a staged rename (index `R` counts as staged, not renamed).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "parse-wrk"
	req.Porcelain = " M modified.txt\n?? untracked.txt\nR  old.txt -> new.txt\n D deleted.txt"
	return nil
}
```
