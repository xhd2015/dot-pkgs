# Scenario

**Feature**: empty gh JSON array returns empty slice without error

```
# no repos for owner
gh repo list carol -> [] -> ListOwned returns []Repo{} nil error
```

## Steps

1. Mock `gh` prints `[]` for owner `carol`.
2. Set `req.Owners` to `["carol"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = []string{"carol"}
	req.GhBin = writeFakeGh(t, `if [ "$1" = "repo" ] && [ "$2" = "list" ] && [ "$3" = "carol" ]; then
  echo '[]'
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`)
	return nil
}```