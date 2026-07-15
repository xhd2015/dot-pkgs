# Scenario

**Feature**: empty tag list yields empty inventory

```
[] -> CollectFromNames -> empty CollectedTags
```

## Steps

1. Set `req.Names` to an empty slice.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{}
	return nil
}
```