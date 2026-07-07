# Scenario

**Feature**: wrk porcelain taxonomy maps lines to four buckets

```
M + ?? + R + D lines -> ParsePorcelainWrk -> Added=1, Changed=1, Renamed=1, Deleted=1
```

## Steps

1. Set `req.Op` to `"parse-wrk"`.
2. Set porcelain with one line per wrk bucket (`??` counts as added).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "parse-wrk"
	req.Porcelain = " M modified.txt\n?? untracked.txt\nR  old.txt -> new.txt\nD  deleted.txt"
	return nil
}
```