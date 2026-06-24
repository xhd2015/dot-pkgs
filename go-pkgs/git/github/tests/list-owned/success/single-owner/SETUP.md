# Scenario

**Feature**: single owner returns two repos mapped and sorted

```
# one owner query
ListOwned owners=[alice] -> gh repo list alice -> 2 repos -> sorted FullName
```

## Steps

1. Load `testdata/repos.json` into mock `gh` for owner `alice`.
2. Set `req.Owners` to `["alice"]`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Owners = []string{"alice"}
	req.GhBin = writeFakeGhFromFixture(t, "testdata/repos.json")
	return nil
}```