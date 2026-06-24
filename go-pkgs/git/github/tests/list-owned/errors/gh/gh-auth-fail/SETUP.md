# Scenario

**Feature**: gh auth failure hints gh auth login

```
# gh exit 4
gh repo list failuser -> exit 4 + auth stderr -> error hints gh auth login
```

## Steps

1. Mock `gh` exits 4 with auth message on stderr.
2. Set `req.Owners` to `["failuser"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Owners = []string{"failuser"}
	req.GhBin = writeFakeGh(t, `if [ "$1" = "repo" ] && [ "$2" = "list" ]; then
  echo 'To authenticate, please run gh auth login' >&2
  exit 4
fi
echo "unexpected args: $*" >&2
exit 1
`)
	return nil
}```