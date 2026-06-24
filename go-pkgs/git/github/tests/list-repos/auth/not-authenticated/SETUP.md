# Scenario

**Feature**: unauthenticated `gh api user` hints `gh auth login`

```
# auth failure
EnsureAuthenticated -> gh api user -> exit 4 -> error contains gh auth login
```

## Steps

1. Mock `gh api user` exits 4 with auth stderr.
2. Leave `req.Owners` empty (inherited nil).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GhBin = writeAuthFailGh(t)
	return nil
}
```