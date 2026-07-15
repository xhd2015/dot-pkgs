# Scenario

**Feature**: root-scope tags have empty `PathPrefix` and `VersionPrefix` of `v`

```
vX.Y.Z[-suffix] -> ParseTagName -> PathPrefix="", VersionPrefix="v"
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