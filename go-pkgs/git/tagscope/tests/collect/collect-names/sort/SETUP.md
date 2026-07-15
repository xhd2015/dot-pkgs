# Scenario

**Feature**: tags within a scope sort newest-first via git version:refname semantics

```
unordered tag names -> CollectFromNames -> Tags newest-first per scope
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Op != "collect-names" {
		t.Fatalf("Op = %q, want collect-names", req.Op)
	}
	return nil
}
```