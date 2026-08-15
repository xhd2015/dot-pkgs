# Scenario

**Feature**: scoped tags carry slash-normalized path prefix before `v`

```
{path/}vX.Y.Z[-suffix] -> ParseTagName -> PathPrefix=path/, VersionPrefix=path/v
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Op != "parse" {
		t.Fatalf("Op = %q, want parse", req.Op)
	}
	return nil
}
```