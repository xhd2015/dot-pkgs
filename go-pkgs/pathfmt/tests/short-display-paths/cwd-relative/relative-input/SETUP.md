# Scenario

**Feature**: relative input under cwd is absolutized then displayed as `"child"`

```
# formatter pipeline
caller path string -> Short -> Abs normalize -> cwd/home rules -> display string
```

## Steps

1. Set `req.Path` to the relative path `"child"` (not absolute).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Path = "child"
	return nil
}```
