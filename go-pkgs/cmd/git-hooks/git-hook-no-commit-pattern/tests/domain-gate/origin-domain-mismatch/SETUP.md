# Scenario

**Feature**: `--origin-domain` mismatch skips the hook silently

```
# hook --origin-domain=other.example.com -> origin is github.com -> skip
hook binary --origin-domain other.example.com -> domain mismatch -> exit 0, no output
```

## Preconditions

- The repository origin is `git@github.com:owner/repo.git`.

## Steps

1. Set `--origin-domain` to a mismatching domain `other.example.com`.
2. Stage a file matching a pattern to prove the hook would run otherwise.
3. Run the hook.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--origin-domain", "other.example.com", "*.go"}
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	return nil
}
```
