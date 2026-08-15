# Scenario

**Feature**: `ListOwned` forwards gh CLI flags from `Options`

```
# flag mapping
IncludeArchived=false -> --no-archived | IncludeForks=false -> --source | Limit -> --limit N
```

## Preconditions

- Mock `gh` captures argv and returns minimal valid JSON.

## Steps

1. Configure `req` option fields per leaf.
2. Install mock `gh` that records `"$*"` and returns one repo.

## Context

- Leaves assert `resp.GhArgv` contains expected flag tokens.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if len(req.Owners) == 0 {
		req.Owners = []string{"optsuser"}
	}
	req.GhBin = writeFakeGh(t, `if [ "$1" = "repo" ] && [ "$2" = "list" ]; then
  echo '[{"name":"one","url":"https://github.com/optsuser/one","description":"","isFork":false,"isArchived":false,"owner":{"login":"optsuser"}}]'
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`)
	return nil
}```