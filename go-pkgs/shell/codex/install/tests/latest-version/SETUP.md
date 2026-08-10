# Scenario

**Feature**: `LatestVersion` — GET npm registry JSON `version` via injectable HTTP

```
GET /@openai/codex/latest (fake) -> JSON {"version":…} -> version string | error
```

## Steps

1. Set `Operation=latest-version`.
2. Leaves set `HTTPMode` / `NPMVersion`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "latest-version"
	return nil
}
```
