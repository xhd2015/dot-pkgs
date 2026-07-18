# Scenario

**Feature**: ApplyLiveness drops dead repo entries from an in-memory index

```
# liveness filter
index with live path + dead path
  -> ApplyLiveness(index)
  -> only entries whose path still has .git remain
```

## Preconditions

- Leaves set `IndexOp` to `liveness`.
- Liveness is FS-based (`.git` present as dir or gitfile); does not rewrite
  `repos.json` unless a later phase wires that in (out of scope for P1 assert).

## Steps

1. Set `IndexOp` to `liveness`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.IndexOp = "liveness"
	req.Universe = "home"
	return nil
}
```
