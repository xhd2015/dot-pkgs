# Scenario

**Feature**: FullName is owner slash name

```
# FullName construction
owner=o name=r -> FullName o/r
```

## Steps

1. Set `req.FullNameOwner` to `o` and `req.FullNameName` to `r`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.FullNameOwner = "o"
	req.FullNameName = "r"
	return nil
}```