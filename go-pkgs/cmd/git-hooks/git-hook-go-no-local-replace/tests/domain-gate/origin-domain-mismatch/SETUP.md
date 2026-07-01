# Scenario

**Feature**: `--origin-domain` mismatch skips the hook silently

```
# hook --origin-domain=other.example.com -> origin is git.xxx.com -> skip
hook binary --origin-domain other.example.com -> domain mismatch -> exit 0, no output
```

## Preconditions

- The repository origin is `git@github.com:owner/repo.git`.

## Steps

1. Set `--origin-domain` to a mismatching domain `other.example.com`.
2. Write a go.mod with a local replace to prove the hook would run otherwise.
3. Run the hook.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--origin-domain", "other.example.com"}
	return nil
}

```
