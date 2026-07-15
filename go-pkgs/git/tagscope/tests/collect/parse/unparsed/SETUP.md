# Scenario

**Feature**: non-matching tag names are rejected by `ParseTagName`

```
invalid tag name -> ParseTagName -> ok=false
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if req.Op != "parse" {
		t.Fatalf("Op = %q, want parse", req.Op)
	}
	return nil
}
```