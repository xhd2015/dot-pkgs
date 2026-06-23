# Scenario

**Feature**: `"~/foo/bar"` expands to `filepath.Join(home, "foo", "bar")`

```
# Expand pipeline
"~/..." -> filepath.Join(home, suffix)
```

## Steps

1. Set `req.Path` to `"~/foo/bar"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Path = "~/foo/bar"
	return nil
}```
