# Scenario

**Feature**: invalid JSON stdout returns decode error

```
# JSON decode failure
gh repo list failuser -> stdout not JSON -> decode error
```

## Steps

1. Mock `gh` prints `not json` to stdout and exits 0.
2. Set `req.Owners` to `["failuser"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Owners = []string{"failuser"}
	req.GhBin = writeFakeGh(t, `if [ "$1" = "repo" ] && [ "$2" = "list" ]; then
  echo 'not json'
  exit 0
fi
echo "unexpected args: $*" >&2
exit 1
`)
	return nil
}```