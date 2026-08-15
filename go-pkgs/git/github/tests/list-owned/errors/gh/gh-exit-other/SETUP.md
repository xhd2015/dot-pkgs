# Scenario

**Feature**: generic gh failure mentions owner and stderr

```
# gh exit 1
gh repo list failuser -> exit 1 + stderr -> error mentions owner and stderr text
```

## Steps

1. Mock `gh` exits 1 with `rate limit exceeded` on stderr.
2. Set `req.Owners` to `["failuser"]`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Owners = []string{"failuser"}
	req.GhBin = writeFakeGh(t, `if [ "$1" = "repo" ] && [ "$2" = "list" ]; then
  echo 'rate limit exceeded' >&2
  exit 1
fi
echo "unexpected args: $*" >&2
exit 1
`)
	return nil
}```