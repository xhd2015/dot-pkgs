# Scenario

**Feature**: empty owners inferred from authenticated login

```
# owner inference
ListRepos Owners=[] -> EnsureAuthenticated login=alice -> gh repo list alice
```

## Steps

1. Mock `gh api user` returns login `alice`.
2. Mock `gh repo list alice` returns two owned repos.
3. Keep `req.Owners` nil (inherited).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GhBin = writeInferOwnerGh(t, "testdata/repos.json")
	return nil
}
```