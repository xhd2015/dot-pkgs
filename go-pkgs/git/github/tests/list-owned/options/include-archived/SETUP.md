# Scenario

**Feature**: IncludeArchived true omits `--no-archived` flag

```
# include archived
IncludeArchived=true -> gh repo list without --no-archived
```

## Steps

1. Set `req.IncludeArchived` to true.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.IncludeArchived = true
	return nil
}```