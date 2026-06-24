# Scenario

**Feature**: `Options` validation rejects invalid owner lists before gh

```
# pre-exec validation
ListOwned opts.Owners invalid -> error (gh never spawned)
```

## Preconditions

- A trap mock `gh` is installed; it must not be executed.

## Steps

1. Install `writeTrapGh` and assign to `req.GhBin`.
2. Leaf `Setup` sets invalid `req.Owners`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GhBin = writeTrapGh(t)
	return nil
}```