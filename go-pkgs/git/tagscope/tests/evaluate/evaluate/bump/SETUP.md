# Scenario

**Feature**: Evaluate plans NextTag when owned files changed

```
scope lineage passes gates + DiffOwnedTrees true -> IncrementTag -> NextTag
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Op != "evaluate" {
		t.Fatalf("Op = %q, want evaluate", req.Op)
	}
	return nil
}
```