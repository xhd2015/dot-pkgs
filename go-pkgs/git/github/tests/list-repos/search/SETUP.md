# Scenario

**Feature**: description and/or code search modes via `gh search`

```
# search branches
ListRepos SearchDescription and/or SearchCode -> gh search repos/code -> matched_by tags
```

## Steps

1. Default owner `alice` when leaf does not set explicit owners.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if len(req.Owners) == 0 {
		req.Owners = []string{"alice"}
	}
	return nil
}
```