# Scenario

**Feature**: `cmd.Run` executes git in a working directory

```
temp repo dir + git args -> Run -> stdout or normalized error
```

## Preconditions

- `git` on PATH (skip otherwise).

## Steps

1. Ensure git availability check runs for every leaf in this branch.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	gitAvailable(t)
	return nil
}
```
