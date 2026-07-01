# Scenario

**Feature**: `--exclude-origin-domain` match skips the hook silently

```
# hook --exclude-origin-domain=github.com -> origin is github.com -> skip
hook binary --exclude-origin-domain github.com -> domain matches exclude -> exit 0, no output
```

## Preconditions

- The repository origin is `git@github.com:owner/repo.git`.

## Steps

1. Set `--exclude-origin-domain` to `github.com` (matching the origin).
2. Run the hook.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--exclude-origin-domain", "github.com"}
	return nil
}

```
