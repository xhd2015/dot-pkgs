# Scenario

**Feature**: blank owner string in Owners slice rejected

```
# invalid owner token
ListOwned owners=[xhd2015,""] -> invalid owner
```

## Steps

1. Set `req.Owners` to `["xhd2015", ""]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Owners = []string{"xhd2015", ""}
	return nil
}```