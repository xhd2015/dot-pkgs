# Scenario

**Feature**: relative real path is Abs+Clean normalized before the mirror key

```
# relative and absolute forms share one mirror entry path
RealPath="proj/nested" -> Abs+Clean -> same MirrorEntryPath as abs form
```

## Steps

1. Set `req.RealPath` to a relative multi-segment path `proj/nested`
   (resolved against process CWD via Abs; path need not exist).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RealPath = "proj/nested"
	return nil
}
```
