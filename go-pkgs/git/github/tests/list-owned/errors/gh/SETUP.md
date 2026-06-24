# Scenario

**Feature**: gh process failures surface as wrapped errors

```
# gh exec failure
ListOwned -> gh exits non-zero or prints invalid JSON -> error
```

## Preconditions

- `req.Owners` contains a valid non-empty owner unless leaf overrides.
- Mock `gh` simulates the failure mode under test.

## Steps

1. Leaf `Setup` installs a failure mock and sets `req.Owners`.

## Context

- Auth failure uses exit code 4 per gh conventions.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Owners) == 0 {
		req.Owners = []string{"failuser"}
	}
	return nil
}```