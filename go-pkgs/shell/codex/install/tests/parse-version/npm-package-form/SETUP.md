# Scenario

**Feature**: parse npm package version form

```
"@openai/codex, 0.122.0" -> ParseVersion -> "0.122.0"
```

## Steps

1. Set `VersionOutput` to `@openai/codex, …` package form.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.VersionOutput = "@openai/codex, 0.122.0"
	return nil
}
```
