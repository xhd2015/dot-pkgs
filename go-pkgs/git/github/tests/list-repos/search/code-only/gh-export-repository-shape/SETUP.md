# Scenario

**Reproduces**: `kool github repo list --search-code doctest` fails with
`decode gh search code JSON: missing repository owner` against real `gh` output.

Real `gh search code --json repository` emits repository objects via gh CLI's
`Repository.MarshalJSON`: `id`, `nameWithOwner`, `url`, `isPrivate`, `isFork`
— not `owner.login`.

## Preconditions

- Mock `gh` returns code-search JSON matching gh CLI export shape (no `owner` field).

## Steps

1. Set `req.SearchCode` to `doctest`.
2. Point `req.GhBin` at mock `gh` serving `testdata/search-code.json`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.SearchCode = "doctest"
	req.GhBin = writeSearchCodeGh(t, fixtureFile(d, "testdata/search-code.json"))
	return nil
}
```