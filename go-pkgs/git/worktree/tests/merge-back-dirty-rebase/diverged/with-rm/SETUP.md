# Scenario

**Feature**: diverged + --rm → existing behavior (requires clean)

```
# Remove=true → IsClean guard still applies, even for diverged
dirty feat -> MergeBack(Remove=true) -> IsClean fails -> error
```

## Steps

1. `Remove=true` set at this level.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Remove = true
	return nil
}
```
