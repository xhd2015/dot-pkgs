# Scenario

**Feature**: `"~"` expands to the absolute home directory

```
# Expand pipeline
"~" -> home directory
```

## Steps

1. Set `req.Path` to `"~"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Path = "~"
	return nil
}```
